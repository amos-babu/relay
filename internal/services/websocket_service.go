package services

import (
	"relay/internal/websocket"
)

type WebSocketService struct{}

func NewWebsocketService() *WebSocketService {
	return &WebSocketService{}
}

func (s *WebSocketService) HandleEvent(
	userID int64,
	event websocket.Event,
) {
	switch {
	case websocket.EventTyping:
	default:
	}
}
