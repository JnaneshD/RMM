package main

import (
	"context"
	"log"
	"sync"
	"time"

	"example.com/test/models"
	"example.com/test/wsheartbeat"
	"github.com/gorilla/websocket"
)

/*
This file should contain all the logic related to the
execution of the task
*/

type JobScheduler struct {
	Jobs []models.Job
	mu   sync.RWMutex
	conn *websocket.Conn
	send chan *models.Job
}

func (jsler *JobScheduler) AddJobImmediate(newjob *models.Job) {
	jsler.mu.Lock()
	jsler.Jobs = append(jsler.Jobs, *newjob)
	jsler.mu.Unlock()
	Execute(newjob)

	// Non blocking send to write channel
	select {
	case jsler.send <- newjob:
	default:
		log.Printf("Send Channel is full, dropping result %s", newjob.Command)
	}
}

func Worker(ctx context.Context, id int, jobQueue <-chan *models.Job, scheduler *JobScheduler) {
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

func (jsler *JobScheduler) writePump() {
	for newjob := range jsler.send {
		jsler.conn.SetWriteDeadline(time.Now().Add(wsheartbeat.WriteWait))
		if err := jsler.conn.WriteJSON(newjob); err != nil {
			log.Printf("WriteJSON errored for job %s: %v", newjob.Command, err)
			return
		}
	}
}
