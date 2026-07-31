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
func (r *ConversationRepository) IsParticipant(ctx context.Context, conversationID int64, userID int64) (bool, error) {
	const query = `
	SELECT EXISTS (
		SELECT 1
		FROM conversation_participants
		WHERE conversation_id = $1
			AND user_id = $2
	);
	`
	var exists bool
	if err := r.db.QueryRowContext(
		ctx,
		query,
		conversationID,
		userID,
	).Scan(
		&exists,
	); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *ConversationRepository) Participants(ctx context.Context, conversationID int64) ([]int64, error) {
	const query = `
	SELECT user_id
	FROM conversation_participants
	WHERE conversation_id = $1
	ORDER BY user_id;
	`
	var participants []int64

	rows, err := r.db.QueryContext(
		ctx,
		query,
		conversationID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var userID int64
		if err := rows.Scan(
			&userID,
		); err != nil {
			return nil, err
		}

		participants = append(participants, userID)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil

}
