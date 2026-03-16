package models

import (
	"time"
)

type Client struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	IsOnline  bool      `json:"is_online"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
