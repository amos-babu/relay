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

func (s *WebSocketService) HandleEvent(userID int64, event websocket.Event) {
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

		//Build the event
		typing := websocket.TypingEvent{
			ConversationID: req.ConversationID,
			UserID:         userID,
		}

		outgoing := websocket.Event{
			Type:    websocket.EventTyping,
			Payload: typing,
		}

		//Marshall the event built
		message, err := json.Marshal(outgoing)
		if err != nil {
			return
		}

		//Find the conversations Participants
		participants, err := s.conversations.Participants(
			context.Background(),
			req.ConversationID,
		)
		if err != nil {
			return
		}

		//Loop through all the Participants
		for _, participantID := range participants {
			if participantID == userID {
				continue
			}

			//Send the event to each user
			s.hub.SendToUser(participantID, message)
		}

	default:
		log.Printf("unknown websocket event: %s", event.Type)
	}
}

func (s *WebSocketService) HandleConnect(userID int64) {

	//Broadcast the event to every user
	s.broadcastPresence(userID, true)

	// log.Printf("user %d connected:", userID)
}
func (s *WebSocketService) HandleDisconnect(userID int64) {
	//If there is still a connection return
	if s.hub.HasConnections(userID) {
		return
	}

	//Broadcast the event to every user
	s.broadcastPresence(userID, false)

	// log.Printf("user %d disconnected:", userID)
}

// Helper function for connect and disconnect events
func (s *WebSocketService) broadcastPresence(userID int64, status bool) {
	//Build the event
	presence := websocket.PresenceEvent{
		UserID: userID,
		Online: status,
	}

	event := websocket.Event{
		Type:    websocket.EventPresence,
		Payload: presence,
	}

	//Marshall the event built
	message, err := json.Marshal(event)
	if err != nil {
		return
	}

	//Broadcast the event to every user
	s.hub.Broadcast(message)
}
