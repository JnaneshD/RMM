package domain

type Job struct {
	ID       uint64    `json:"id"`
	ClientID string    `json:"client_id"`
	Command  string    `json:"command"`
	Status   JobStatus `json:"status"`
	Output   string    `json:"output"`
}

type JobStatus int

const (
	WAIT JobStatus = iota
	RUNNING
	FINISHED
	FAILED
)
