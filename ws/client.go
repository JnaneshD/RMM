package ws

import (
	"example.com/test/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	conn *websocket.Conn
	Send chan models.Job
}

func NewClient(clientID string, conn *websocket.Conn) *Client {
	return &Client{
		ID:   clientID,
		conn: conn,
		Send: make(chan models.Job),
	}
}
