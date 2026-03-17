package ws

import (
	"fmt"
	"sync"

	"example.com/test/models"
)

type Hub struct {
	clients     map[string]*Client
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	stop        chan struct{}
	Client_Jobs map[*Client][]models.Job
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		stop:        make(chan struct{}),
		Client_Jobs: make(map[*Client][]models.Job),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			fmt.Printf("Agent %s got connected\n", client.ID)
		case client := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, client.ID)
			h.mu.Unlock()
			close(client.Send)
			fmt.Printf("Agent %s got disconnected\n", client.ID)
		case <-h.stop:
			return
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Clients() map[string]*Client {
	return h.clients
}

func (h *Hub) GetClient(id string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, exists := h.clients[id]
	return client, exists
}

func (h *Hub) Stop() {
	close(h.stop)
}

func (h *Hub) AddJobToClient(job models.Job, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Client_Jobs[client] = append(h.Client_Jobs[client], job)
}
