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

	OnConnect    func(userID int64)
	OnDisconnect func(userID int64)
	onEvent      func(int64, Event)
}

func NewHandler(
	hub *Hub,
	OnConnect func(userID int64),
	OnDisconnect func(userID int64),
	onEvent func(int64, Event),
) *Handler {
	return &Handler{
		hub:          hub,
		OnConnect:    OnConnect,
		OnDisconnect: OnDisconnect,
		onEvent:      onEvent,
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
		Send:   make(chan []byte, 256),

		OnConnect:    h.OnConnect,
		OnDisconnect: h.OnDisconnect,
		OnEvent:      h.onEvent,
	}

	//Register
	h.hub.Register(client)

	//Calling onConnect event functions
	if client.OnConnect != nil {
		client.OnConnect(client.UserID)
	}

	go client.writePump()

	defer func() {
		//CleanUp
		h.hub.Unregister(client)

		if client.OnDisconnect != nil {
			client.OnDisconnect(client.UserID)
		}

		close(client.Send)
		conn.Close()

	}()

	client.readPump()

}
