package realtime

import (
	"example.com/test/internal/domain"
	"github.com/gorilla/websocket"
)

type ActiveClient struct {
	ID   string
	conn *websocket.Conn
	Send chan domain.Job
}

func NewClient(clientID string, conn *websocket.Conn) *ActiveClient {
	return &ActiveClient{
		ID:   clientID,
		conn: conn,
	}
}
