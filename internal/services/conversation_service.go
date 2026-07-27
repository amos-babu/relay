package services

import (
	"context"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
)

type ConversationService struct {
	conversations repositories.ConversationRepository
	users         repositories.UserRepository
}

func NewConversationService(
	conversations repositories.ConversationRepository,
	users repositories.UserRepository,
) *ConversationService {
	return &ConversationService{
		conversations: conversations,
		users:         users,
	}
}
func (s *ConversationService) Create(
	ctx context.Context,
	creatorID int64,
	recipientID int64,
) (*models.Conversation, error) {
	// A user cannot start a conversation with themselves.
	if creatorID == recipientID {
		return nil, domain.ErrCannotMessageYourself
	}

	//Ensure the recipient exists
	if _, err := s.users.GetByID(ctx, recipientID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	// Delegate persistence to the repository.
	return s.conversations.Create(ctx, creatorID, recipientID)

}

func (s *ConversationService) ListForUser(
	ctx context.Context,
	userID int64,
) ([]*models.Conversation, error) {

	return s.conversations.ListForUser(ctx, userID)
}
