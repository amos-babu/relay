package services

import (
	"relay/internal/repositories"
	"relay/internal/websocket"
)

type WebSocketService struct {
	conversations repositories.ConversationRepository
	hub           *websocket.Hub
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
