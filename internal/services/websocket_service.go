package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"relay/internal/domain"
	"relay/internal/events"
	"relay/internal/repositories"
	"relay/internal/websocket"
	"time"
)

type PresenceTracker interface {
	Connect(ctx context.Context, userID int64, connectionID string) (bool, error)
	Heartbeat(ctx context.Context, userID int64, connectionID string) error
	Disconnect(ctx context.Context, userID int64, connectionID string) (bool, error)
}

type WebSocketService struct {
	conversations   repositories.ConversationRepository
	messageService  *MessageService
	hub             *websocket.Hub
	publisher       EventPublisher
	presenceTracker PresenceTracker
}

func NewWebsocketService(
	conversations repositories.ConversationRepository,
	messageService *MessageService,
	hub *websocket.Hub,
	publisher EventPublisher,
	presenceTracker PresenceTracker,
) *WebSocketService {
	return &WebSocketService{
		conversations:   conversations,
		messageService:  messageService,
		hub:             hub,
		publisher:       publisher,
		presenceTracker: presenceTracker,
	}
}

func (s *WebSocketService) HandleEvent(userID int64, event websocket.Event) {
	switch event.Type {

	case websocket.EventTyping:
		s.HandleTyping(userID, event)

	case websocket.EventReadReceipt:
		s.HandleReadReceipt(userID, event)

	default:
		log.Printf("unknown websocket event: %s", event.Type)
	}
}

func (s *WebSocketService) HandleTyping(userID int64, event websocket.Event) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		log.Printf("failed to marshal typing payload: %v", err)
		return
	}

	var req websocket.TypingRequest

	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("failed to unmarshal typing request: %v", err)
		return
	}

	ctx := context.Background()

	// Verify that the sender belongs to the conversation.
	ok, err := s.conversations.IsParticipant(
		ctx,
		req.ConversationID,
		userID,
	)
	if err != nil {
		log.Printf("failed to check typing participant: %v", err)
		return
	}

	if !ok {
		log.Printf(
			"typing event rejected: user %d is not a participant in conversation %d",
			userID,
			req.ConversationID,
		)
		return
	}

	// Build the typing event.
	typing := websocket.TypingEvent{
		ConversationID: req.ConversationID,
		UserID:         userID,
	}

	outgoing := websocket.Event{
		Type:    websocket.EventTyping,
		Payload: typing,
	}

	// Find all participants in the conversation.
	participants, err := s.conversations.Participants(
		ctx,
		req.ConversationID,
	)
	if err != nil {
		log.Printf(
			"failed to get participants for conversation %d: %v",
			req.ConversationID,
			err,
		)
		return
	}

	// Exclude the sender from the recipients.
	recipients := make([]int64, 0, len(participants))

	for _, participantID := range participants {
		if participantID == userID {
			continue
		}

		recipients = append(recipients, participantID)
	}

	// Publish through Redis so other Relay instances can deliver the event.
	broadcast := websocket.BroadcastEvent{
		RecipientIDs: recipients,
		Event:        outgoing,
	}

	if err := s.publisher.Publish(ctx, broadcast); err != nil {
		log.Printf(
			"failed to publish typing event for user %d in conversation %d: %v",
			userID,
			req.ConversationID,
			err,
		)
	}
}

func (s *WebSocketService) HandleReadReceipt(userID int64, event websocket.Event) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return
	}

	var req websocket.ReadReceiptRequest

	if err := json.Unmarshal(payload, &req); err != nil {
		return
	}

	//Massage Service Call
	_, err = s.messageService.MarkAsRead(
		context.Background(),
		req.MessageID,
		req.ConversationID,
		userID,
	)

	if err != nil {
		if errors.Is(err, domain.ErrNotConversationParticipant) {
			log.Printf(
				"read receipt rejected: user %d is not a participant in conversation %d",
				userID,
				req.ConversationID,
			)
			return
		}

		if errors.Is(err, domain.ErrMessageNotFound) {
			log.Printf(
				"read receipt rejected: message %d does not belong to conversation %d",
				req.MessageID,
				req.ConversationID,
			)
			return
		}

		log.Printf(
			"failed to mark message %d as read: %v",
			req.MessageID,
			err,
		)
		return

	}

}

func (s *WebSocketService) HandleConnect(userID int64, connectionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	online, err := s.presenceTracker.Connect(ctx, userID, connectionID)
	if err != nil {
		log.Printf(
			"Failed to register presence for user %d, connection %s: %v",
			userID,
			connectionID,
			err,
		)
		return
	}

	// Announce online only when this is the user's first active connection.
	if online {
		s.broadcastPresence(userID, true)
	}
}

func (s *WebSocketService) HandleDisconnect(userID int64, connectionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offline, err := s.presenceTracker.Disconnect(ctx, userID, connectionID)
	if err != nil {
		log.Printf(
			"Failed to unregister presence for user %d, connection %s: %v",
			userID,
			connectionID,
			err,
		)
		return
	}

	// Announce offline only when no active connections remain.
	if offline {
		s.broadcastPresence(userID, false)
	}
}

// Showing Offline and Online
func (s *WebSocketService) broadcastPresence(userID int64, status bool) {
	ctx := context.Background()

	// Find all conversations belonging to this user.
	conversations, err := s.conversations.ListForUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list conversations for user %d: %v", userID, err)
		return
	}

	// Collect unique participants, excluding the user whose status changed.
	recipientSet := make(map[int64]struct{})

	for _, conversation := range conversations {
		participants, err := s.conversations.Participants(ctx, conversation.ID)
		if err != nil {
			log.Printf(
				"failed to get participants for conversation %d: %v",
				conversation.ID,
				err,
			)
			return
		}

		for _, participantID := range participants {
			if participantID == userID {
				continue
			}

			recipientSet[participantID] = struct{}{}
		}
	}

	// Convert the set into a slice for the Redis broadcast event.
	recipients := make([]int64, 0, len(recipientSet))

	for recipientID := range recipientSet {
		recipients = append(recipients, recipientID)
	}

	// Build the presence event.
	presence := websocket.PresenceEvent{
		UserID: userID,
		Online: status,
	}

	event := websocket.Event{
		Type:    websocket.EventPresence,
		Payload: presence,
	}

	broadcast := websocket.BroadcastEvent{
		RecipientIDs: recipients,
		Event:        event,
	}

	// Publish the event so every Relay instance can deliver it locally.
	if err := s.publisher.Publish(ctx, broadcast); err != nil {
		log.Printf(
			"failed to publish presence event for user %d: %v",
			userID,
			err,
		)
	}

}

func (s *WebSocketService) HandleHeartbeat(
	userID int64,
	connectionID string,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	err := s.presenceTracker.Heartbeat(ctx, userID, connectionID)
	if err == nil {
		return
	}

	// Only attempt recovery when the connection record is missing.
	if errors.Is(err, events.ErrConnectionNotRegistered) {
		log.Printf(
			"Presence record missing for user %d, connection %s; attempting recovery",
			userID,
			connectionID,
		)

		online, reconnectErr := s.presenceTracker.Connect(
			ctx,
			userID,
			connectionID,
		)
		if reconnectErr != nil {
			log.Printf(
				"Failed to recover presence for user %d, connection %s: %v",
				userID,
				connectionID,
				reconnectErr,
			)
			return
		}

		if online {
			s.broadcastPresence(userID, true)
		}

		log.Printf(
			"Recovered presence for user %d, connection %s",
			userID,
			connectionID,
		)
		return
	}

	// Redis may be unavailable. Don't assume recovery succeeded.
	log.Printf(
		"Failed to refresh presence for user %d, connection %s: %v",
		userID,
		connectionID,
		err,
	)
}
