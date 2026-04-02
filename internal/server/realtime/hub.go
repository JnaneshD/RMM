package realtime

import (
	"log"
	"sync"
)

type Hub struct {
	clients    map[string]*ActiveClient
	register   chan *ActiveClient
	unregister chan *ActiveClient
	mu         sync.RWMutex
	stop       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*ActiveClient),
		register:   make(chan *ActiveClient),
		unregister: make(chan *ActiveClient),
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

func (h *Hub) Register(client *ActiveClient) {
	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()
	log.Printf("Agent %s got connected\n", client.ID)
}

func (h *Hub) Unregister(client *ActiveClient) {
	h.mu.Lock()
	delete(h.clients, client.ID)
	h.mu.Unlock()
	if client.Send != nil {
		close(client.Send)
	}
	log.Printf("Agent %s got disconnected\n", client.ID)
}

func (h *Hub) GetClient(id string) (*ActiveClient, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, exists := h.clients[id]
	return client, exists
}

func (h *Hub) Stop() {
	close(h.stop)
}
