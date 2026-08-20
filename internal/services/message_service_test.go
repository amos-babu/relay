package services

import (
	"context"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"testing"
	"time"
)

type fakeConversationRepository struct {
	IsParticipantFunc func(ctx context.Context, conversationID, userID int64) (bool, error)
}

type fakeMessageRepository struct {
	CreateFunc func(ctx context.Context, message *models.Message) error
}

func (f *fakeConversationRepository) IsParticipant(
	ctx context.Context,
	conversationID int64,
	userID int64,
) (bool, error) {
	if f.IsParticipantFunc != nil {
		return f.IsParticipantFunc(ctx, conversationID, userID)
	}
	return false, nil
}

func (f *fakeMessageRepository) Create(ctx context.Context, message *models.Message) error {
	return f.CreateFunc(ctx, message)
}

func (f *fakeConversationRepository) Create(
	ctx context.Context,
	creatorID int64,
	recipientID int64,
) (*models.Conversation, error) {
	panic("Create should not be called")
}

func (f *fakeConversationRepository) ListForUser(
	ctx context.Context,
	userID int64,
) ([]*models.Conversation, error) {
	panic("ListForUser should not be called")
}

func (f *fakeConversationRepository) FindDirectConversation(
	ctx context.Context,
	user1ID int64,
	user2ID int64,
) (*models.Conversation, error) {
	panic("FindDirectConversation should not be called")
}

func (f *fakeConversationRepository) Participants(
	ctx context.Context,
	conversationID int64,
) ([]int64, error) {
	panic("Participants should not be called")
}

func (f *fakeMessageRepository) ListForConversation(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
	panic("ListForConversation should not be called")
}
func (f *fakeMessageRepository) MarkAsRead(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error) {
	panic("MarkAsRead should not be called")
}
func (f *fakeMessageRepository) GetReadReceipts(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
	panic("GetReadReceipts should not be called")
}

func TestMessageService_Send_EmptyMessage(t *testing.T) {
	service := &MessageService{}

	_, err := service.Send(
		context.Background(),
		1,
		1,
		"  ",
	)

	if err != domain.ErrEmptyMessage {
		t.Fatalf(
			"expected ErrEmptyMessage, got %v",
			err,
		)
	}
}
func TestMessageService_Send_NotParticipant(t *testing.T) {
	//Arrange: Setup fake repository behavior
	fakeRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, nil
		},
	}

	service := &MessageService{
		conversations: fakeRepo,
	}

	//Act
	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	//Assert
	if !errors.Is(err, domain.ErrNotConversationParticipant) {
		t.Fatalf(
			"expected ErrNotConversationParticipant, got %v",
			err,
		)
	}
}

func TestMessageService_Send_ParticipantCheckError(t *testing.T) {
	//Arrange: Setup fake repository behavior
	expectedErr := errors.New("database error")
	fakeRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeRepo,
	}

	//Act
	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	//Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestMessageService_Send_CreateError(t *testing.T) {
	//Arrange: Setup fake repository behavior
	expectedErr := errors.New("failed to save message")

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		CreateFunc: func(ctx context.Context, message *models.Message) error {
			return expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	//Act
	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	//Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestMessageService_Send_Success(t *testing.T) {
	//Arrange: Setup fake repository behavior
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		CreateFunc: func(ctx context.Context, message *models.Message) error {
			return nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	//Act
	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	//Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

}
