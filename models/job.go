package models

type Job struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	ClientID string `gorm:"index" json:"client_id"`
	Command  string `json:"command"`
	Status   string `json:"status" default:"pending"`
}
