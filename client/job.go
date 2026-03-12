// Lets create a job type with necessary fields
package main

import (
	"os/exec"
	"sync/atomic"
)

type JobStatus int

const (
	WAIT JobStatus = iota
	RUNNING
	FINISHED
	FAILED
)

type Job struct {
	JobId   int
	command string
	Output  string
	Status  JobStatus
}

var jobId atomic.Int64

func NewJob(command string) *Job {
	return &Job{
		JobId:   int(jobId.Add(1)),
		command: command,
		Status:  WAIT,
	}
}

func (job *Job) Execute() {
	// Now we need to execute the job in the client
	cmd := exec.Command("cmd", "/C", job.command)
	stdout, err := cmd.Output()

	if err != nil {
		job.Output = err.Error()
		job.Status = FAILED
	} else {
		job.Status = FINISHED
		job.Output = string(stdout)
	}
}
