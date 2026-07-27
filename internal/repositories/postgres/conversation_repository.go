package postgres

import (
	"context"
	"database/sql"
	"relay/internal/models"
	"relay/internal/repositories"
)

type ConversationRepository struct {
	db *sql.DB
}

var _ repositories.ConversationRepository = (*ConversationRepository)(nil)

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

func (r *ConversationRepository) Create(ctx context.Context, creatorID, recipientID int64) (*models.Conversation, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	const conversationQuery = `
	INSERT INTO conversations
	DEFAULT VALUES
	RETURNING id, created_at;
	`

	// insert conversation
	conversation := &models.Conversation{}
	err = tx.QueryRowContext(
		ctx,
		conversationQuery,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// insert creator
	const participantsQuery = `
	INSERT INTO conversation_participants (
		conversation_id,
		user_id
	)
	VALUES ($1, $2);
	`

	if _, err := tx.ExecContext(
		ctx,
		participantsQuery,
		conversation.ID,
		creatorID,
	); err != nil {
		return nil, err
	}

	// insert recipient
	_, err = tx.ExecContext(
		ctx,
		participantsQuery,
		conversation.ID,
		recipientID,
	)

	if err != nil {
		return nil, err
	}

	// commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return conversation, nil
}
func (r *ConversationRepository) GetByID(ctx context.Context, id int64) (*models.Conversation, error) {
	return nil, nil
}
func (r *ConversationRepository) ListForUser(ctx context.Context, userID int64) ([]*models.Conversation, error) {
	return nil, nil
}
