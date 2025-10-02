package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"threadbound/internal/models"
	"threadbound/internal/service"
)

// Job represents a book generation job
type Job struct {
	ID         string
	Status     JobStatus
	Config     *models.BookConfig
	Result     *service.GenerateResult
	Error      error
	CreatedAt  time.Time
	UpdatedAt  time.Time
	cancelFunc func()
}

// JobManager manages async job processing
type JobManager struct {
	jobs  map[string]*Job
	mutex sync.RWMutex
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

// CreateJob creates a new job and starts processing it asynchronously
func (jm *JobManager) CreateJob(config *models.BookConfig) string {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	jobID := uuid.New().String()
	job := &Job{
		ID:        jobID,
		Status:    JobStatusPending,
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	jm.jobs[jobID] = job

	// Start processing in background
	go jm.processJob(jobID)

	return jobID
}

// processJob processes a job asynchronously
func (jm *JobManager) processJob(jobID string) {
	jm.mutex.Lock()
	job, exists := jm.jobs[jobID]
	if !exists {
		jm.mutex.Unlock()
		return
	}
	job.Status = JobStatusRunning
	job.UpdatedAt = time.Now()
	jm.mutex.Unlock()

	// Create generator service
	genService := service.NewGeneratorService(job.Config)

	// Execute generation
	result, err := genService.Generate()

	// Update job with result
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job.UpdatedAt = time.Now()
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err
	} else {
		job.Status = JobStatusCompleted
		job.Result = result
	}
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(jobID string) (*Job, error) {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all jobs
func (jm *JobManager) ListJobs() []*Job {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()

	jobs := make([]*Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// DeleteJob removes a job from the manager
func (jm *JobManager) DeleteJob(jobID string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	if _, exists := jm.jobs[jobID]; !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	delete(jm.jobs, jobID)
	return nil
}
