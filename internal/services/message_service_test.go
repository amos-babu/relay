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
	ParticipantsFunc  func(ctx context.Context, conversationID int64) ([]int64, error)
}

type fakeMessageRepository struct {
	CreateFunc     func(ctx context.Context, message *models.Message) error
	MarkAsReadFunc func(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error)
}

type fakeHub struct {
	sentTo   []int64
	payloads [][]byte
}

// FakeConversationRepo Methods
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

// FakeMessageRepo Methods
func (f *fakeMessageRepository) Create(ctx context.Context, message *models.Message) error {
	return f.CreateFunc(ctx, message)
}
func (f *fakeMessageRepository) ListForConversation(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
	panic("ListForConversation should not be called")
}
func (f *fakeMessageRepository) MarkAsRead(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error) {
	if f.MarkAsReadFunc != nil {
		return f.MarkAsReadFunc(ctx, messageID, conversationID, userID)
	}
	panic("MarkAsRead should not be called")
}
func (f *fakeMessageRepository) GetReadReceipts(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
	panic("GetReadReceipts should not be called")
}

// FakeHub Interface Methods
func (f *fakeHub) SendToUser(userID int64, message []byte) {
	f.sentTo = append(f.sentTo, userID)
	f.payloads = append(f.payloads, message)
}

// Tests
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
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
			return []int64{1, 2}, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		CreateFunc: func(ctx context.Context, message *models.Message) error {
			message.ID = 24
			message.CreatedAt = time.Now()

			return nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		hub:           fakeHub,
	}

	// Act
	_, err := service.Send(
		context.Background(),
		1,
		1,
		"Hello, world",
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 1. Verify Recipients
	expectedRecipients := []int64{1, 2}
	if len(fakeHub.sentTo) != len(expectedRecipients) {
		t.Fatalf("expected %d deliveries, got %d", len(expectedRecipients), len(fakeHub.sentTo))
	}

	for i, expectedID := range expectedRecipients {
		if fakeHub.sentTo[i] != expectedID {
			t.Fatalf("expected recipient index %d to be %d, got %d", i, expectedID, fakeHub.sentTo[i])
		}
	}

	// 2. Verify WebSocket Event Payload safely
	if len(fakeHub.payloads) == 0 {
		t.Fatal("expected broadcast payload, got none")
	}

	var event websocket.Event
	if err := json.Unmarshal(fakeHub.payloads[0], &event); err != nil {
		t.Fatalf("failed to decode websocket event: %v", err)
	}

	if event.Type != websocket.EventMessage {
		t.Fatalf(
			"expected event type %q, got %q",
			websocket.EventMessage,
			event.Type,
		)
	}

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		t.Fatalf("failed to re-marshal payload: %v", err)
	}

	//Unmarshall the messageevent
	var messageEvent MessageEvent

	if err := json.Unmarshal(payloadBytes, &messageEvent); err != nil {
		t.Fatalf("failed to decode message payload: %v", err)
	}

	//Verifying the actual messages
	if messageEvent.ID == 0 {
		t.Fatal("expected message ID to be set")
	}

	if messageEvent.ConversationID != 1 {
		t.Fatalf(
			"expected conversation ID 1, got %d",
			messageEvent.ConversationID,
		)
	}

	if messageEvent.SenderID != 1 {
		t.Fatalf(
			"expected sender ID 1, got %d",
			messageEvent.SenderID,
		)
	}

	if messageEvent.Content != "Hello, world" {
		t.Fatalf(
			"expected content %q, got %q",
			"Hello, world",
			messageEvent.Content,
		)
	}

}

func TestMessageService_Send_ParticipantsError(t *testing.T) {
	//Arrage
	expectedErr := errors.New("failed to fetch participants")
	fakeHub := &fakeHub{}
	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
			return nil, expectedErr
		},
	}

	fakeMessageRepository := &fakeMessageRepository{
		CreateFunc: func(ctx context.Context, message *models.Message) error {
			message.ID = 1
			message.CreatedAt = time.Now()
			return nil
		},
	}

	service := &MessageService{
		messages:      fakeMessageRepository,
		conversations: fakeConversationRepository,
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
	//Check if err is same as our error
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	//Verify no messages were sent over websocket
	if len(fakeHub.sentTo) != 0 {
		t.Fatalf(
			"expected no messages to be sent, got %d",
			len(fakeHub.sentTo),
		)
	}

}

func TestMessageService_MarkAsRead_NotParticipant(t *testing.T) {
	//Arrange
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
	}

	//Act
	_, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	//Assert
	if !errors.Is(err, domain.ErrNotConversationParticipant) {
		t.Fatalf(
			"expected ErrNotConversationParticipant, got %v",
			err,
		)
	}
}

func TestMessageService_MarkAsRead_ParticipantCheckError(t *testing.T) {
	//Arrange
	expectedErr := errors.New("database error")
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
	}

	//Act
	_, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	//Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestMessageService_MarkAsRead_MessageNotFound(t *testing.T) {
	//Arrange
	expectedErr := domain.ErrMessageNotFound

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return time.Time{}, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	//Act
	_, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	//Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

}
func TestMessageService_MarkAsRead_RepositoryError(t *testing.T) {
	//Arrange
	expectedErr := errors.New("database error")

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return time.Time{}, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	//Act
	_, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	//Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
func TestMessageService_MarkAsRead_Success(t *testing.T) {
	// Arrange
	expectedReadAt := time.Now()

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(
			ctx context.Context,
			conversationID int64,
			userID int64,
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
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return expectedReadAt, nil
		},
	}

	fakeHub := &fakeHub{}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		hub:           fakeHub,
	}

	// Act
	readAt, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if readAt.IsZero() {
		t.Fatal("expected readAt to be set")
	}

	if !readAt.Equal(expectedReadAt) {
		t.Fatalf(
			"expected readAt %v, got %v",
			expectedReadAt,
			readAt,
		)
	}
}

func TestMessageService_MarkAsRead_ParticipantsError(t *testing.T) {
	//Arrage
	expectedErr := errors.New("failed to fetch participants")
	expectedReadAt := time.Now()
	fakeHub := &fakeHub{}
	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
			return nil, expectedErr
		},
	}

	fakeMessageRepository := &fakeMessageRepository{
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return expectedReadAt, nil
		},
	}

	service := &MessageService{
		messages:      fakeMessageRepository,
		conversations: fakeConversationRepository,
		hub:           fakeHub,
	}

	//Act
	_, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	//Assert
	//Check if err is same as our error
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	//Verify no messages were sent over websocket
	if len(fakeHub.sentTo) != 0 {
		t.Fatalf(
			"expected no messages to be sent, got %d",
			len(fakeHub.sentTo),
		)
	}

}

func TestMessageService_MarkAsRead_BroadcastsReadReceipt(t *testing.T) {
	// Arrange
	fakeHub := &fakeHub{}
	expectedReadAt := time.Now()

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(
			ctx context.Context,
			conversationID int64,
			userID int64,
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
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return expectedReadAt, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		hub:           fakeHub,
	}

	// Act
	readAt, err := service.MarkAsRead(
		context.Background(),
		1,
		1,
		1,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !readAt.Equal(expectedReadAt) {
		t.Fatalf(
			"expected readAt %v, got %v",
			expectedReadAt,
			readAt,
		)
	}

	// Verify only the other participant receives the event.
	expectedRecipients := []int64{2}

	if len(fakeHub.sentTo) != len(expectedRecipients) {
		t.Fatalf(
			"expected %d deliveries, got %d",
			len(expectedRecipients),
			len(fakeHub.sentTo),
		)
	}

	for i, expectedID := range expectedRecipients {
		if fakeHub.sentTo[i] != expectedID {
			t.Fatalf(
				"expected recipient index %d to be %d, got %d",
				i,
				expectedID,
				fakeHub.sentTo[i],
			)
		}
	}

	// Verify a payload was sent.
	if len(fakeHub.payloads) == 0 {
		t.Fatal("expected broadcast payload, got none")
	}

	// Decode the WebSocket event.
	var event websocket.Event

	if err := json.Unmarshal(fakeHub.payloads[0], &event); err != nil {
		t.Fatalf(
			"failed to decode websocket event: %v",
			err,
		)
	}

	// Verify event type.
	if event.Type != websocket.EventReadReceipt {
		t.Fatalf(
			"expected event type %q, got %q",
			websocket.EventReadReceipt,
			event.Type,
		)
	}

	// Decode the read receipt payload.
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		t.Fatalf(
			"failed to re-marshal payload: %v",
			err,
		)
	}

	var readReceipt websocket.ReadReceiptEvent

	if err := json.Unmarshal(payloadBytes, &readReceipt); err != nil {
		t.Fatalf(
			"failed to decode read receipt: %v",
			err,
		)
	}

	// Verify read receipt contents.
	if readReceipt.MessageID != 1 {
		t.Fatalf(
			"expected message ID 1, got %d",
			readReceipt.MessageID,
		)
	}

	if readReceipt.ConversationID != 1 {
		t.Fatalf(
			"expected conversation ID 1, got %d",
			readReceipt.ConversationID,
		)
	}

	if readReceipt.UserID != 1 {
		t.Fatalf(
			"expected user ID 1, got %d",
			readReceipt.UserID,
		)
	}

	if !readReceipt.ReadAt.Equal(expectedReadAt) {
		t.Fatalf(
			"expected readAt %v, got %v",
			expectedReadAt,
			readReceipt.ReadAt,
		)
	}
}
