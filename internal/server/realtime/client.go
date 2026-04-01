package realtime

import (
	"example.com/test/internal/domain"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID          string
	conn        *websocket.Conn
	Send        chan domain.Job
	Fingerprint string
	HostName    string
}

type ClientResponse struct {
	ID       string
	HostName string
}

func NewClient(clientID string, fingerprint string, hostname string) *Client {
	return &Client{
		ID:          clientID,
		Fingerprint: fingerprint,
		HostName:    hostname,
	}
}

func (cl *Client) UpdateClient(conn *websocket.Conn) {
	cl.conn = conn
	cl.Send = make(chan domain.Job)
}
