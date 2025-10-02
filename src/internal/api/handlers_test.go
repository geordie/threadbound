package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestHealthEndpoint(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestGenerateEndpoint(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	genReq := GenerateRequest{
		DatabasePath:  "/path/to/test.db",
		Title:         "Test Book",
		IncludeImages: true,
	}

	body, err := json.Marshal(genReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	var response GenerateResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.JobID == "" {
		t.Error("Expected job_id to be set")
	}

	if response.Status != JobStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", response.Status)
	}
}

func TestGenerateEndpointMissingDatabasePath(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	genReq := GenerateRequest{
		Title: "Test Book",
	}

	body, err := json.Marshal(genReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "database_path is required" {
		t.Errorf("Expected error 'database_path is required', got '%s'", response.Error)
	}
}

func TestGetJobStatus(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// First create a job
	genReq := GenerateRequest{
		DatabasePath:  "/path/to/test.db",
		Title:         "Test Book",
		IncludeImages: true,
	}

	body, _ := json.Marshal(genReq)
	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var genResponse GenerateResponse
	json.NewDecoder(w.Body).Decode(&genResponse)

	// Give the job a moment to potentially start
	time.Sleep(100 * time.Millisecond)

	// Now check the job status
	req = httptest.NewRequest("GET", "/api/jobs/"+genResponse.JobID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var statusResponse JobStatusResponse
	err := json.NewDecoder(w.Body).Decode(&statusResponse)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if statusResponse.JobID != genResponse.JobID {
		t.Errorf("Expected job_id '%s', got '%s'", genResponse.JobID, statusResponse.JobID)
	}

	// Status should be pending, running, completed, or failed
	validStatuses := []JobStatus{JobStatusPending, JobStatusRunning, JobStatusCompleted, JobStatusFailed}
	validStatus := false
	for _, status := range validStatuses {
		if statusResponse.Status == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		t.Errorf("Invalid status: %s", statusResponse.Status)
	}
}

func TestGetJobStatusNotFound(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/api/jobs/non-existent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestListJobs(t *testing.T) {
	handler := NewHandler()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Create a couple of jobs
	for i := 0; i < 2; i++ {
		genReq := GenerateRequest{
			DatabasePath:  "/path/to/test.db",
			Title:         "Test Book",
			IncludeImages: true,
		}

		body, _ := json.Marshal(genReq)
		req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// List all jobs
	req := httptest.NewRequest("GET", "/api/jobs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var jobs []JobStatusResponse
	err := json.NewDecoder(w.Body).Decode(&jobs)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(jobs) < 2 {
		t.Errorf("Expected at least 2 jobs, got %d", len(jobs))
	}
}
