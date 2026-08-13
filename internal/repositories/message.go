package repositories

import (
	"context"
	"relay/internal/domain"
	"relay/internal/models"
	"time"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	ListForConversation(ctx context.Context, conversationID int64) ([]*models.Message, error)
	MarkAsRead(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error)
	GetReadReceipts(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error)
}
