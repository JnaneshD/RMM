package ws

import (
	"github.com/gorilla/websocket"
)

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
