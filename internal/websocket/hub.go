package websocket

import (
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	clients map[int64]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
	}

	h.clients[client.UserID][client] = true
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	delete(clients, client)

	if len(clients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:

		default:

		}
	}

}
