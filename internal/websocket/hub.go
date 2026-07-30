package websocket

import (
	"sync"
)

type Hub struct {
	mu sync.RWMutex

	clients map[int64]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.UserID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client.UserID)
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[userID]
	if !ok {
		return
	}

	select {
	case client.Send <- message:

	default:

	}
}
