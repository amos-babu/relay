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

	resp := MessageEvent{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	for _, userID := range participants {
		s.hub.SendToUser(userID, payload)
	}

	return message, nil
}

func (s *MessageService) ListForConversation(ctx context.Context, conversationID int64, userID int64) ([]*models.Message, error) {
	//Check if sender is a participant in this conversation
	ok, err := s.conversations.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotConversationParticipant
	}

	//Fetch the messages
	messages, err := s.messages.ListForConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
