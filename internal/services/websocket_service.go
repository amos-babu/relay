package services

import (
	"context"
	"encoding/json"
	"log"
	"relay/internal/repositories"
	"relay/internal/websocket"
)

type WebSocketService struct {
	conversations repositories.ConversationRepository
	hub           *websocket.Hub
}

func NewWebsocketService(conversations repositories.ConversationRepository, hub *websocket.Hub) *WebSocketService {
	return &WebSocketService{
		conversations: conversations,
		hub:           hub,
	}
}

func (s *WebSocketService) HandleEvent(
	userID int64,
	event websocket.Event,
) {
	switch event.Type {

	case websocket.EventTyping:
		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return
		}

		var req websocket.TypingRequest

		if err := json.Unmarshal(payload, &req); err != nil {
			return
		}

		ok, err := s.conversations.IsParticipant(
			context.Background(),
			req.ConversationID,
			userID,
		)

		if err != nil {
			return
		}

		if !ok {
			return
		}

		typing := websocket.TypingEvent{
			ConversationID: req.ConversationID,
			UserID:         userID,
		}

		outgoing := websocket.Event{
			Type:    websocket.EventTyping,
			Payload: typing,
		}

		message, err := json.Marshal(outgoing)
		if err != nil {
			return
		}

		participants, err := s.conversations.Participants(
			context.Background(),
			req.ConversationID,
		)
		if err != nil {
			return
		}

		for _, participantID := range participants {
			if participantID == userID {
				continue
			}

			s.hub.SendToUser(participantID, message)
		}

	default:
		log.Printf("unknown websocket event: %s", event.Type)
	}
}
