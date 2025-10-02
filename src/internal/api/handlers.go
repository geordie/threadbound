package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"threadbound/internal/models"
)

// Handler manages API request handling
type Handler struct {
	jobManager *JobManager
}

// NewHandler creates a new API handler
func NewHandler() *Handler {
	return &Handler{
		jobManager: NewJobManager(),
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/generate", h.handleGenerate).Methods("POST")
	r.HandleFunc("/api/jobs/{job_id}", h.handleGetJobStatus).Methods("GET")
	r.HandleFunc("/api/jobs", h.handleListJobs).Methods("GET")
	r.HandleFunc("/api/health", h.handleHealth).Methods("GET")
}

// handleGenerate handles POST /api/generate
func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if req.DatabasePath == "" {
		respondError(w, http.StatusBadRequest, "database_path is required", nil)
		return
	}

	// Create book config from request
	config := &models.BookConfig{
		DatabasePath:    req.DatabasePath,
		AttachmentsPath: req.AttachmentsPath,
		OutputPath:      req.OutputPath,
		Title:           req.Title,
		Author:          req.Author,
		PageWidth:       req.PageWidth,
		PageHeight:      req.PageHeight,
		IncludeImages:   req.IncludeImages,
		IncludePreviews: true,
		ContactNames:    req.ContactNames,
		MyName:          req.MyName,
	}

	// Set defaults
	if config.AttachmentsPath == "" {
		config.AttachmentsPath = "Attachments"
	}
	if config.OutputPath == "" {
		config.OutputPath = "book.tex"
	}
	if config.Title == "" {
		config.Title = "Our Messages"
	}
	if config.PageWidth == "" {
		config.PageWidth = "5.5in"
	}
	if config.PageHeight == "" {
		config.PageHeight = "8.5in"
	}

	// Create and start job
	jobID := h.jobManager.CreateJob(config)

	// Return response
	resp := GenerateResponse{
		JobID:     jobID,
		Status:    JobStatusPending,
		Message:   "Job created successfully",
		CreatedAt: time.Now(),
	}

	respondJSON(w, http.StatusAccepted, resp)
}

// handleGetJobStatus handles GET /api/jobs/{job_id}
func (h *Handler) handleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	job, err := h.jobManager.GetJob(jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Job not found", err)
		return
	}

	resp := JobStatusResponse{
		JobID:     job.ID,
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}

	if job.Error != nil {
		resp.Error = job.Error.Error()
	}

	if job.Result != nil {
		resp.OutputPath = job.Result.OutputPath
		if job.Result.Stats != nil {
			resp.Stats = &JobStats{
				TotalMessages:   job.Result.Stats.TotalMessages,
				TextMessages:    job.Result.Stats.TextMessages,
				TotalContacts:   job.Result.Stats.TotalContacts,
				AttachmentCount: job.Result.Stats.AttachmentCount,
				StartDate:       job.Result.Stats.StartDate,
				EndDate:         job.Result.Stats.EndDate,
			}
		}
	}

	switch job.Status {
	case JobStatusPending:
		resp.Message = "Job is pending"
	case JobStatusRunning:
		resp.Message = "Job is running"
	case JobStatusCompleted:
		resp.Message = "Job completed successfully"
	case JobStatusFailed:
		resp.Message = "Job failed"
	}

	respondJSON(w, http.StatusOK, resp)
}

// handleListJobs handles GET /api/jobs
func (h *Handler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.jobManager.ListJobs()

	responses := make([]JobStatusResponse, 0, len(jobs))
	for _, job := range jobs {
		resp := JobStatusResponse{
			JobID:     job.ID,
			Status:    job.Status,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		}

		if job.Error != nil {
			resp.Error = job.Error.Error()
		}

		if job.Result != nil {
			resp.OutputPath = job.Result.OutputPath
		}

		responses = append(responses, resp)
	}

	respondJSON(w, http.StatusOK, responses)
}

// handleHealth handles GET /api/health
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := ErrorResponse{
		Error:   message,
		Message: message,
	}
	if err != nil {
		resp.Message = err.Error()
	}
	respondJSON(w, statusCode, resp)
}
