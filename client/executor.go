package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

/*
This file should contain all the logic related to the
execution of the task
*/

type JobScheduler struct {
	Jobs []Job
	mu   sync.RWMutex
	conn *websocket.Conn
}

func (jsler *JobScheduler) AddJobImmediate(command string) {
	newjob := NewJob(command)
	jsler.mu.Lock()
	jsler.Jobs = append(jsler.Jobs, *newjob)
	jsler.mu.Unlock()
	newjob.Execute()
	jsler.conn.WriteMessage(websocket.TextMessage, []byte(newjob.Output))
}
