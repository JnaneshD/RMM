package logic

import (
	"fmt"

	models "example.com/test/models"
)

type Hub struct {
	models.Hub
}

func NewHub() *Hub {
	return &Hub{
		Hub: models.Hub{
			Clients:    make(map[string]*models.Client),
			Register:   make(chan *models.Client),
			Unregister: make(chan *models.Client),
		},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Hub.Register:
			h.Hub.Mu.Lock()
			h.Hub.Clients[client.ID] = client
			h.Hub.Mu.Unlock()
			fmt.Printf("Agent %s got connected\n", client.ID)
		case client := <-h.Hub.Unregister:
			h.Hub.Mu.Lock()
			delete(h.Hub.Clients, client.ID)
			h.Hub.Mu.Unlock()
			close(client.Send)
			fmt.Printf("Agent %s got disconnected\n", client.ID)
		case <-h.Stop:
			return
		}
	}
}

func (h *Hub) Register(client *models.Client) {
	h.Hub.Register <- client
}

func (h *Hub) Unregister(client *models.Client) {
	h.Hub.Unregister <- client
}

func (h *Hub) Clients() map[string]*models.Client {
	return h.Hub.Clients
}
