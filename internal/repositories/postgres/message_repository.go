package postgres

import (
	"context"
	"database/sql"
	"relay/internal/models"
	"relay/internal/repositories"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

var _ repositories.MessageRepository = (*MessageRepository)(nil)

func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	return nil
}
func (r *MessageRepository) ListForConversation(ctx context.Context, conversationID int64) ([]*models.Message, error) {
	return nil, nil
}
