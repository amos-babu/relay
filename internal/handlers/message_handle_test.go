package handlers

import (
	"context"
	"relay/internal/models"
	"relay/internal/services"
	"time"
)

type fakeMessageService struct {
	sendFunc func(
		ctx context.Context,
		conversationID int64,
		senderID int64,
		content string,
	) (*models.Message, error)
}

func (f *fakeMessageService) Send(
	ctx context.Context,
	conversationID int64,
	senderID int64,
	content string,
) (*models.Message, error) {
	return f.sendFunc(ctx, conversationID, senderID, content)
}

func (f *fakeMessageService) ListForConversation(
	ctx context.Context,
	conversationID int64,
	userID int64,
	limit int,
	before *int64,
) (*services.ConversationMessages, error) {
	panic("not implemented")
}

func (f *fakeMessageService) MarkAsRead(
	ctx context.Context,
	messageID int64,
	conversationID int64,
	userID int64,
) (time.Time, error) {
	panic("not implemented")
}
