package handlers

import (
	"context"
	"relay/internal/models"
	"relay/internal/services"
	"time"
)

type MessageService interface {
	Send(
		ctx context.Context,
		conversationID int64,
		senderID int64,
		content string,
	) (*models.Message, error)

	ListForConversation(
		ctx context.Context,
		conversationID int64,
		userID int64,
		limit int,
		before *int64,
	) (*services.ConversationMessages, error)

	MarkAsRead(
		ctx context.Context,
		messageID int64,
		conversationID int64,
		userID int64,
	) (time.Time, error)
}
