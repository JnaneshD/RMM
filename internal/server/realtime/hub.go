package realtime

import (
	"fmt"
	"log"
	"sync"
)

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	stop       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Agent %s got connected\n", client.ID)
		case client := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, client.ID)
			h.mu.Unlock()
			close(client.Send)
			log.Printf("Agent %s got disconnected\n", client.ID)
		case <-h.stop:
			return
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()
	log.Printf("Agent %s got connected\n", client.ID)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	delete(h.clients, client.ID)
	h.mu.Unlock()
	if client.Send != nil {
		close(client.Send)
	}
	log.Printf("Agent %s got disconnected\n", client.ID)
}

func (h *Hub) GetClient(id string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	fmt.Println(h.clients)
	client, exists := h.clients[id]
	return client, exists
}

func (h *Hub) ClientIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) Stop() {
	close(h.stop)
}
