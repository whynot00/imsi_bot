package handler

import (
	"sync"

	"github.com/google/uuid"
	"github.com/whynot00/imsi-bot/internal/service"
)

// JobStatus represents the current state of an upload job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

// Job tracks a single background import.
type Job struct {
	ID     string                `json:"id"`
	Status JobStatus             `json:"status"`
	Result *service.ImportResult `json:"result,omitempty"`
	Error  string                `json:"error,omitempty"`
}

// JobStore is a thread-safe in-memory store for upload jobs.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

func (s *JobStore) Create() *Job {
	j := &Job{
		ID:     uuid.NewString(),
		Status: JobStatusPending,
	}
	s.mu.Lock()
	s.jobs[j.ID] = j
	s.mu.Unlock()
	return j
}

func (s *JobStore) Get(id string) *Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs[id]
}

func (s *JobStore) SetProcessing(id string) {
	s.mu.Lock()
	if j, ok := s.jobs[id]; ok {
		j.Status = JobStatusProcessing
	}
	s.mu.Unlock()
}

func (s *JobStore) SetDone(id string, result *service.ImportResult) {
	s.mu.Lock()
	if j, ok := s.jobs[id]; ok {
		j.Status = JobStatusDone
		j.Result = result
	}
	s.mu.Unlock()
}

func (s *JobStore) SetFailed(id string, err string) {
	s.mu.Lock()
	if j, ok := s.jobs[id]; ok {
		j.Status = JobStatusFailed
		j.Error = err
	}
	s.mu.Unlock()
}
