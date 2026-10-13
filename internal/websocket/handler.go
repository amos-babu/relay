package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"relay/internal/middleware"
	"relay/internal/response"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub

	OnConnect    func(userID int64, connectionID string)
	OnDisconnect func(userID int64, connectionID string)
	OnHeartbeat  func(userID int64, connectionID string)

	onEvent func(userID int64, event Event)
}

func NewHandler(
	hub *Hub,
	onConnect func(userID int64, connectionID string),
	onDisconnect func(userID int64, connectionID string),
	onHeartbeat func(userID int64, connectionID string),
	onEvent func(userID int64, event Event),
) *Handler {
	return &Handler{
		hub:          hub,
		OnConnect:    onConnect,
		OnDisconnect: onDisconnect,
		OnHeartbeat:  onHeartbeat,
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

	connectionID, err := generateConnectionID()
	if err != nil {
		log.Printf("Failed to generate WebSocket connection ID: %v", err)
		conn.Close()
		return
	}

	//Create the Client after authentication
	client := &Client{
		UserID:       userID,
		ConnectionID: connectionID,
		Conn:         conn,
		Send:         make(chan []byte, 256),

		OnEvent:      h.onEvent,
		OnConnect:    h.OnConnect,
		OnDisconnect: h.OnDisconnect,
		OnHeartbeat:  h.OnHeartbeat,
	}

	//Register
	h.hub.Register(client)

	//Calling onConnect event functions
	if client.OnConnect != nil {
		client.OnConnect(client.UserID, client.ConnectionID)
	}

	go client.writePump()

	defer func() {
		//CleanUp
		h.hub.Unregister(client)

		if client.OnDisconnect != nil {
			client.OnDisconnect(client.UserID, client.ConnectionID)
		}

		close(client.Send)
		conn.Close()

	}()

	client.readPump()

}

func generateConnectionID() (string, error) {
	id := make([]byte, 16)

	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	return hex.EncodeToString(id), nil
}
