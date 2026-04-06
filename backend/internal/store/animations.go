package store

import (
	"sync"
	"time"
)

type AnimationSegment struct {
	Index         int
	Status        string
	Duration      int
	Prompt        string
	SourceJobID   string
	VideoURL      string
	LastFramePath string
	Error         string
}

type AnimationJob struct {
	ID                string
	Status            string
	Provider          string
	Model             string
	Prompt            string
	DurationSeconds   int
	AspectRatio       string
	LeadImage         string
	Seed              *int64
	SegmentCount      int
	CompletedSegments int
	CurrentSegment    int
	VideoURL          string
	Error             string
	Segments          []AnimationSegment
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type AnimationStore struct {
	mu   sync.RWMutex
	jobs map[string]*AnimationJob
}

func NewAnimationStore() *AnimationStore {
	return &AnimationStore{jobs: map[string]*AnimationJob{}}
}

func (s *AnimationStore) Put(job *AnimationJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = cloneAnimationJob(job)
}

func (s *AnimationStore) Get(id string) (*AnimationJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok || job == nil {
		return nil, false
	}
	return cloneAnimationJob(job), true
}

func (s *AnimationStore) Update(id string, fn func(job *AnimationJob)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok || job == nil {
		return false
	}
	fn(job)
	job.UpdatedAt = time.Now()
	return true
}

func cloneAnimationJob(in *AnimationJob) *AnimationJob {
	if in == nil {
		return nil
	}
	out := *in
	if len(in.Segments) > 0 {
		out.Segments = append([]AnimationSegment(nil), in.Segments...)
	}
	if in.Seed != nil {
		seed := *in.Seed
		out.Seed = &seed
	}
	return &out
}
