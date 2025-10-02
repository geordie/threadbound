package api

import "time"

// JobStatus represents the status of a generation job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// GenerateRequest represents a request to generate a book
type GenerateRequest struct {
	DatabasePath    string            `json:"database_path"`
	AttachmentsPath string            `json:"attachments_path,omitempty"`
	OutputPath      string            `json:"output_path,omitempty"`
	Title           string            `json:"title,omitempty"`
	Author          string            `json:"author,omitempty"`
	PageWidth       string            `json:"page_width,omitempty"`
	PageHeight      string            `json:"page_height,omitempty"`
	IncludeImages   bool              `json:"include_images"`
	ContactNames    map[string]string `json:"contact_names,omitempty"`
	MyName          string            `json:"my_name,omitempty"`
}

// GenerateResponse represents the response to a generate request
type GenerateResponse struct {
	JobID     string    `json:"job_id"`
	Status    JobStatus `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// JobStatusResponse represents the status of a job
type JobStatusResponse struct {
	JobID      string    `json:"job_id"`
	Status     JobStatus `json:"status"`
	Message    string    `json:"message,omitempty"`
	Error      string    `json:"error,omitempty"`
	OutputPath string    `json:"output_path,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Stats      *JobStats `json:"stats,omitempty"`
}

// JobStats contains statistics about the generated book
type JobStats struct {
	TotalMessages   int       `json:"total_messages"`
	TextMessages    int       `json:"text_messages"`
	TotalContacts   int       `json:"total_contacts"`
	AttachmentCount int       `json:"attachment_count"`
	StartDate       time.Time `json:"start_date,omitempty"`
	EndDate         time.Time `json:"end_date,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
