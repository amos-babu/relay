package repositories

import (
	"context"
	"relay/internal/models"
)

type ConversationRepository interface {
	Create(ctx context.Context, creatorID, recipientID int64) (*models.Conversation, error)
	ListForUser(ctx context.Context, userID int64) ([]*models.Conversation, error)
	FindDirectConversation(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error)
	IsParticipant(ctx context.Context, conversationID int64, userID int64) (bool, error)
	// GetByID(ctx context.Context, id int64) (*models.Conversation, error)
}
