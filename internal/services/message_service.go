package services

import (
	"context"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
	"strings"
)

type MessageService struct {
	messages      repositories.MessageRepository
	conversations repositories.ConversationRepository
}

func NewMessageService(messages repositories.MessageRepository, conversations repositories.ConversationRepository) *MessageService {
	return &MessageService{
		messages:      messages,
		conversations: conversations,
	}
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
		return nil, domain.ErrUnauthorizedParticipant
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

	return message, nil
}
