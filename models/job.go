package models

type Job struct {
	ID       uint64    `gorm:"primaryKey" json:"id"`
	ClientID string    `gorm:"index" json:"client_id"`
	Command  string    `json:"command"`
	Status   JobStatus `gorm:"default:0" json:"status"`
	Output   string    `json:"output"`
}
type JobStatus int

const (
	WAIT JobStatus = iota
	RUNNING
	FINISHED
	FAILED
)
