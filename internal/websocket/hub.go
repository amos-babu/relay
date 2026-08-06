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

// Broadcast to a user the message
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

// Broadcast to a user the message
func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- message:

			default:
			}
		}
	}
}

// Helper method to check whether a User still has other connections
func (h *Hub) HasConnections(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	if !ok {
		return false
	}

	return len(clients) > 0
}
