package store

import (
	"sync"
	"time"
)

type VideoJob struct {
	ID              string
	Status          string // queued|running|succeeded|failed
	VideoURL        string
	Error           string
	Prompt          string
	DurationSeconds int
	AspectRatio     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*VideoJob
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: map[string]*VideoJob{}}
}

func (s *JobStore) Put(job *VideoJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

func (s *JobStore) Get(id string) (*VideoJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	return j, ok
}

func (s *JobStore) Update(id string, fn func(job *VideoJob)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return false
	}
	fn(j)
	return true
}
