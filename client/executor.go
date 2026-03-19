package main

import (
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
	jsler.conn.WriteJSON(newjob)
	//jsler.conn.WriteMessage(websocket.TextMessage, []byte(newjob.Output))
}
