package repositories

import (
	"context"
	"relay/internal/models"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	ListForConversation(ctx context.Context, conversationID int64) ([]*models.Message, error)
	MarkAsRead(ctx context.Context, messageID int64, userID int64) error
}
