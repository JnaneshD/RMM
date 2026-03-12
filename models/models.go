package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.RWMutex
	Stop       chan struct{}
}

type Client struct {
	ID   string
	conn *websocket.Conn
	Send chan string
}

func NewClient(clientID string, conn *websocket.Conn) *Client {
	return &Client{
		ID:   clientID,
		conn: conn,
		Send: make(chan string),
	}
}

type Job struct {
	Command string `json:"command" binding:"required"`
}
