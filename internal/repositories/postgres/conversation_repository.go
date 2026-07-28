package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/domain"
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
func (r *ConversationRepository) ListForUser(ctx context.Context, userID int64) ([]*models.Conversation, error) {
	const query = `
	SELECT
		c.id,
		c.created_at
	FROM conversations c
	INNER JOIN conversation_participants cp
		ON cp.conversation_id = c.id
	WHERE cp.user_id = $1
	ORDER BY c.created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var conversations []*models.Conversation

	for rows.Next() {
		conversation := &models.Conversation{}

		if err := rows.Scan(
			&conversation.ID,
			&conversation.CreatedAt,
		); err != nil {
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}
func (r *ConversationRepository) FindDirectConversation(ctx context.Context, user1ID, user2ID int64) (*models.Conversation, error) {
	const query = `
	SELECT
		c.id,
		c.created_at
	FROM conversations c
	JOIN conversation_participants cp
		ON cp.conversation_id = c.id
	WHERE cp.user_id IN ($1, $2)
	GROUP BY c.id, c.created_at
	HAVING COUNT(DISTINCT cp.user_id) = 2;
	`

	conversation := &models.Conversation{}

	if err := r.db.QueryRowContext(
		ctx,
		query,
		user1ID,
		user2ID,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConversationNotFound
		}
		return nil, err
	}

	return conversation, nil
}
func (r *ConversationRepository) GetByID(ctx context.Context, id int64) (*models.Conversation, error) {
	return nil, nil
}
