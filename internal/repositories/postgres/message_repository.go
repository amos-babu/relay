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
	const query = `
	INSERT INTO messages (
		conversation_id,
		sender_id,
		content
	)
	VALUES (
		$1,
		$2,
		$3
	)
	RETURNING id, created_at;
	`
	if err := r.db.QueryRowContext(
		ctx,
		query,
		message.ConversationID,
		message.SenderID,
		message.Content,
	).Scan(
		&message.ID,
		&message.CreatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (r *MessageRepository) ListForConversation(ctx context.Context, conversationID int64) ([]*models.Message, error) {
	const query = `
	SELECT
		id,
		conversation_id,
		sender_id,
		content,
		created_at
	FROM messages
	WHERE conversation_id = $1
	ORDER BY created_at ASC, id ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []*models.Message

	for rows.Next() {
		message := &models.Message{}

		if err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID int64, userID int64) error {
	const query = `
	INSERT INTO message_reads (message_id, user_id)
	VALUES ($1, $2)
	ON CONFLICT (message_id, user_id)
	DO UPDATE SET read_at = NOW();
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		messageID,
		userID,
	)

	return err
}
