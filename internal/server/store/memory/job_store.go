package memory

import (
	"sort"
	"sync"

	"example.com/test/internal/domain"
)

type JobStore struct {
	mu           sync.RWMutex
	nextID       uint64
	jobsByClient map[string]map[uint64]*domain.Job
}

func NewJobStore() *JobStore {
	return &JobStore{
		jobsByClient: make(map[string]map[uint64]*domain.Job),
	}
}

func (s *JobStore) Create(clientID, command string) domain.Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	job := domain.Job{
		ID:       s.nextID,
		ClientID: clientID,
		Command:  command,
		Status:   domain.WAIT,
	}
	s.ensureClientJobsLocked(clientID)
	s.jobsByClient[clientID][job.ID] = &job
	return job
}

func (s *JobStore) Update(job domain.Job) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientJobs, ok := s.jobsByClient[job.ClientID]
	if !ok {
		return false
	}
	if _, ok := clientJobs[job.ID]; !ok {
		return false
	}
	clientJobs[job.ID] = &job
	return true
}

func (s *JobStore) Snapshot() map[string][]domain.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string][]domain.Job, len(s.jobsByClient))
	for clientID, jobs := range s.jobsByClient {
		list := make([]domain.Job, 0, len(jobs))
		for _, job := range jobs {
			list = append(list, *job)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].ID < list[j].ID
		})
		snapshot[clientID] = list
	}
	return snapshot
}

func (s *JobStore) ensureClientJobsLocked(clientID string) {
	if s.jobsByClient[clientID] == nil {
		s.jobsByClient[clientID] = make(map[uint64]*domain.Job)
	}
}
