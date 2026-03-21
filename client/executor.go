package main

import (
	"context"
	"log"
	"sync"

	"example.com/test/models"
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
}

func (jsler *JobScheduler) AddJobImmediate(newjob *models.Job) {
	jsler.mu.Lock()
	jsler.Jobs = append(jsler.Jobs, *newjob)
	jsler.mu.Unlock()
	Execute(newjob)

	// Handle the collisions
	jsler.mu.Lock()
	err := jsler.conn.WriteJSON(newjob)
	jsler.mu.Unlock()
	if err != nil {
		log.Printf("WriteJSON errored for job %s , %v", newjob.Command, err)
	}
	//jsler.conn.WriteMessage(websocket.TextMessage, []byte(newjob.Output))
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
