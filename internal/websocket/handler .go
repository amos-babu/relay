package websocket

import (
	"log"
	"net/http"
	"relay/internal/middleware"
	"relay/internal/response"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// upgrades that HTTP requests into a WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Get the authenticated User from Auth Middleware
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "unathorized")
		return
	}

	//upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	//Create the Client after authentication
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte),
	}

	//Register
	h.hub.Register(client)

	//Testing
	log.Printf("User %d connected", userID)

	defer func() {
		//CleanUp
		h.hub.Unregister(client)
		conn.Close()

		//Testing
		log.Printf("User %d disconnected", userID)
	}()

}
