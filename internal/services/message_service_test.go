package services

import (
	"context"
	"encoding/json"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/websocket"
	"testing"
	"time"
)

type fakeConversationRepository struct {
	IsParticipantFunc func(ctx context.Context, conversationID, userID int64) (bool, error)

	ParticipantsFunc func(
		ctx context.Context,
		conversationID int64,
	) ([]int64, error)
}

type fakeHub struct {
	sentTo   []int64
	payloads [][]byte
}

func (f *fakeHub) SendToUser(userID int64, message []byte) {
	f.sentTo = append(f.sentTo, userID)
	f.payloads = append(f.payloads, message)
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

func (f *fakeConversationRepository) Participants(
	ctx context.Context,
	conversationID int64,
) ([]int64, error) {
	if f.ParticipantsFunc != nil {
		return f.ParticipantsFunc(ctx, conversationID)
	}
	panic("Participants should not be called")
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
		IsParticipantFunc: func(
			ctx context.Context,
			conversationID, userID int64,
		) (bool, error) {
			return true, nil
		},

		ParticipantsFunc: func(
			ctx context.Context,
			conversationID int64,
		) ([]int64, error) {
			return []int64{1, 2}, nil
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

		ParticipantsFunc: func(
			ctx context.Context,
			conversationID int64,
		) ([]int64, error) {
			return []int64{1, 2}, nil
		},
	}

	var createdMessage *models.Message

	fakeMessageRepo := &fakeMessageRepository{
		CreateFunc: func(ctx context.Context, message *models.Message) error {
			createdMessage = message
			return nil
		},
	}

	fakeHub := &fakeHub{}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		hub:           fakeHub,
	}

	//Act
	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	if createdMessage == nil {
		t.Fatalf("expected message to be created")
	}

	if createdMessage.ConversationID != 1 {
		t.Fatalf(
			"expected conversationId 1, got id %v",
			createdMessage.ConversationID,
		)
	}

	if createdMessage.SenderID != 1 {
		t.Fatalf(
			"expected sender ID 1, got %d",
			createdMessage.SenderID,
		)
	}

	if createdMessage.Content != "Hello, world" {
		t.Fatalf(
			"expected content %q, got %q",
			"Hello, world",
			createdMessage.Content,
		)
	}

	//Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

}

func TestMessageService_Send_BroadcastsMessage(t *testing.T) {
	// Arrange
	fakeHub := &fakeHub{}

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},

		ParticipantsFunc: func(
			ctx context.Context,
			conversationID int64,
		) ([]int64, error) {
			return []int64{1, 2}, nil
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
		hub:           fakeHub,
	}

	//Act
	_, err := service.Send(
		context.Background(),
		1,
		1,
		"Hello, world",
	)

	//Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeHub.sentTo) != 2 {
		t.Fatalf(
			"expected 2 deliveries, got %d",
			len(fakeHub.sentTo),
		)
	}

	if fakeHub.sentTo[0] != 1 {
		t.Fatalf("expected first recipient 1, got %d", fakeHub.sentTo[0])
	}

	if fakeHub.sentTo[1] != 2 {
		t.Fatalf("expected second recipient 2, got %d", fakeHub.sentTo[1])
	}

	var event websocket.Event

	err = json.Unmarshal(fakeHub.payloads[0], &event)
	if err != nil {
		t.Fatalf("failed to decode websocket event: %v", err)
	}

	//Verify the event
	if event.Type != websocket.EventMessage {
		t.Fatalf(
			"expected event type %q, got %q",
			websocket.EventMessage,
			event.Type,
		)
	}

}
