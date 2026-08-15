package services

import (
	"context"
	"encoding/json"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
	"relay/internal/websocket"
	"strings"
	"time"
)

type MessageService struct {
	messages      repositories.MessageRepository
	conversations repositories.ConversationRepository
	hub           *websocket.Hub
}

func NewMessageService(messages repositories.MessageRepository, conversations repositories.ConversationRepository, hub *websocket.Hub) *MessageService {
	return &MessageService{
		messages:      messages,
		conversations: conversations,
		hub:           hub,
	}
}

type MessageEvent struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConversationMessages struct {
	Messages   []*models.Message
	Reads      []*domain.MessageRead
	NextCursor *int64
	HasMore    bool
}

func (s *MessageService) Send(ctx context.Context, conversationID int64, senderID int64, content string) (*models.Message, error) {
	//Validate the content
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, domain.ErrEmptyMessage
	}

	//Check if sender is a participant in this conversation
	ok, err := s.conversations.IsParticipant(ctx, conversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotConversationParticipant
	}

	//Build the model
	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
	}

	//Save
	if err := s.messages.Create(ctx, message); err != nil {
		return nil, err
	}

	//Check the other Participant in the conversation
	participants, err := s.conversations.Participants(
		ctx,
		message.ConversationID,
	)
	if err != nil {
		return nil, err
	}

	//Build the messageEvent response
	resp := MessageEvent{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}

	//Build the event
	event := websocket.Event{
		Type:    websocket.EventMessage,
		Payload: resp,
	}

	//Marshall the event
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	//Send Each message event to the websocket hub
	for _, userID := range participants {
		s.hub.SendToUser(userID, payload)
	}

	return message, nil
}

func (s *MessageService) ListForConversation(ctx context.Context, conversationID int64, userID int64, limit int, before *int64) (*ConversationMessages, error) {
	//Check if sender is a participant in this conversation
	ok, err := s.conversations.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotConversationParticipant
	}

	//Fetch the messages
	messages, err := s.messages.ListForConversation(ctx, conversationID, before, limit)
	if err != nil {
		return nil, err
	}

	//Checking if we have received the extra message
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	var nextCursor *int64

	if hasMore && len(messages) > 0 {
		cursor := messages[len(messages)-1].ID
		nextCursor = &cursor
	}

	//Fetch all read receipts for this conversation
	reads, err := s.messages.GetReadReceipts(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	return &ConversationMessages{
		Messages:   messages,
		Reads:      reads,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *MessageService) MarkAsRead(
	ctx context.Context,
	messageID int64,
	conversationID int64,
	userID int64,
) (time.Time, error) {

	//Check if user is a participant in this conversation
	ok, err := s.conversations.IsParticipant(
		ctx,
		conversationID,
		userID,
	)
	if err != nil {
		return time.Time{}, err
	}

	if !ok {
		return time.Time{}, domain.ErrNotConversationParticipant
	}

	//Mark the message as read
	readAt, err := s.messages.MarkAsRead(
		ctx,
		messageID,
		conversationID,
		userID,
	)
	if err != nil {
		return time.Time{}, err
	}

	//Get conversation participants
	participants, err := s.conversations.Participants(ctx, conversationID)
	if err != nil {
		return time.Time{}, err
	}

	//Build the read receipt event
	readReceipt := websocket.ReadReceiptEvent{
		MessageID:      messageID,
		ConversationID: conversationID,
		UserID:         userID,
		ReadAt:         readAt,
	}

	event := websocket.Event{
		Type:    websocket.EventReadReceipt,
		Payload: readReceipt,
	}

	//Marshall event
	payload, err := json.Marshal(event)
	if err != nil {
		return time.Time{}, err
	}

	//Notify the other participants
	for _, participantID := range participants {
		if participantID == userID {
			continue
		}

		s.hub.SendToUser(participantID, payload)
	}

	return readAt, nil
}
