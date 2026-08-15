package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
	"time"
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

func (r *MessageRepository) ListForConversation(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
	const query = `
	SELECT
		id,
		conversation_id,
		sender_id,
		content,
		created_at
	FROM messages
	WHERE conversation_id = $1
		AND ($2::bigint IS NULL OR id < $2)
	ORDER BY id DESC
	LIMIT $3 + 1;
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, before, limit)
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

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error) {
	const query = `
	INSERT INTO message_reads (message_id, user_id)
	SELECT id, $3
	FROM messages
	WHERE id = $1
		AND conversation_id = $2
	ON CONFLICT (message_id, user_id)
	DO UPDATE SET read_at = NOW()
	RETURNING read_at;
	`
	var readAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		query,
		messageID,
		conversationID,
		userID,
	).Scan(&readAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, domain.ErrMessageNotFound
		}
		return time.Time{}, err
	}

	return readAt, nil
}

func (r *MessageRepository) GetReadReceipts(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
	const query = `
		SELECT 
			mr.message_id,
			mr.user_id,
			mr.read_at
		FROM message_reads mr
		INNER JOIN messages m
			ON m.id = mr.message_id
		WHERE m.conversation_id = $1
		ORDER BY mr.message_id, mr.read_at;
		`
	rows, err := r.db.QueryContext(
		ctx,
		query,
		conversationID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var receipts []*domain.MessageRead

	for rows.Next() {
		receipt := &domain.MessageRead{}

		if err := rows.Scan(
			&receipt.MessageID,
			&receipt.UserID,
			&receipt.ReadAt,
		); err != nil {
			return nil, err
		}

		receipts = append(receipts, receipt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return receipts, nil
}
