package repositories

import (
	"context"
	"relay/internal/models"
)

type ConversationRepository interface {
	Create(ctx context.Context, creatorID, recipientID int64) (*models.Conversation, error)
	GetByID(ctx context.Context, id int64) (*models.Conversation, error)
	ListForUser(ctx context.Context, userID int64) ([]*models.Conversation, error)
}
