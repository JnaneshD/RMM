package runtime

import (
	"context"
	"log"
	"sync"
	"time"

	"example.com/test/internal/domain"
	"example.com/test/internal/heartbeat"
	"github.com/gorilla/websocket"
)

type JobScheduler struct {
	Jobs     []domain.Job
	mu       sync.RWMutex
	conn     *websocket.Conn
	send     chan *domain.Job
	executor Executor
}

func NewJobScheduler(conn *websocket.Conn, executor Executor) *JobScheduler {
	return &JobScheduler{
		Jobs:     make([]domain.Job, 0),
		conn:     conn,
		send:     make(chan *domain.Job, 10),
		executor: executor,
	}
}

func (s *JobScheduler) AddJobImmediate(newJob *domain.Job) {
	s.mu.Lock()
	s.Jobs = append(s.Jobs, *newJob)
	s.mu.Unlock()

	ExecuteJob(newJob, s.executor)

	select {
	case s.send <- newJob:
	default:
		log.Printf("Send channel is full, dropping result %s", newJob.Command)
	}
}

func Worker(ctx context.Context, id int, jobQueue <-chan *domain.Job, scheduler *JobScheduler) {
	log.Printf("Worker %d started", id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d: context cancelled successfully", id)
			return
		case job, ok := <-jobQueue:
			if !ok {
				log.Printf("Worker %d, job queue closed, cancelling", id)
				return
			}
			log.Printf("Worker %d, executing the job %s", id, job.Command)
			scheduler.AddJobImmediate(job)
		}
	}
}

func (s *JobScheduler) WritePump() {
	for newJob := range s.send {
		s.conn.SetWriteDeadline(time.Now().Add(heartbeat.WriteWait))
		if err := s.conn.WriteJSON(newJob); err != nil {
			log.Printf("WriteJSON errored for job %s: %v", newJob.Command, err)
			return
		}
	}
}
