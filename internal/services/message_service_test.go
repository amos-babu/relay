package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/websocket"
)

type fakeConversationRepository struct {
	IsParticipantFunc func(ctx context.Context, conversationID, userID int64) (bool, error)
	ParticipantsFunc  func(ctx context.Context, conversationID int64) ([]int64, error)
}

type fakeMessageRepository struct {
	CreateFunc              func(ctx context.Context, message *models.Message) error
	MarkAsReadFunc          func(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error)
	ListForConversationFunc func(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error)
	GetReadReceiptsFunc     func(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error)
}

type fakeEventPublisher struct {
	publishFunc func(ctx context.Context, event websocket.BroadcastEvent) error
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
	if f.ListForConversationFunc != nil {
		return f.ListForConversationFunc(ctx, conversationID, before, limit)
	}
	panic("ListForConversation should not be called")
}

func (f *fakeMessageRepository) MarkAsRead(ctx context.Context, messageID int64, conversationID int64, userID int64) (time.Time, error) {
	if f.MarkAsReadFunc != nil {
		return f.MarkAsReadFunc(ctx, messageID, conversationID, userID)
	}
	panic("MarkAsRead should not be called")
}

func (f *fakeMessageRepository) GetReadReceipts(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
	if f.GetReadReceiptsFunc != nil {
		return f.GetReadReceiptsFunc(ctx, conversationID)
	}
	panic("GetReadReceipts should not be called")
}

// FakeEventPublisher Interface Methods
func (f *fakeEventPublisher) Publish(ctx context.Context, event websocket.BroadcastEvent) error {
	if f.publishFunc != nil {
		return f.publishFunc(ctx, event)
	}
	panic("Publish should not be called")
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

	if !errors.Is(err, domain.ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage, got %v", err)
	}
}

func TestMessageService_Send_NotParticipant(t *testing.T) {
	fakeRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, nil
		},
	}

	service := &MessageService{
		conversations: fakeRepo,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	if !errors.Is(err, domain.ErrNotConversationParticipant) {
		t.Fatalf("expected ErrNotConversationParticipant, got %v", err)
	}
}

func TestMessageService_Send_ParticipantCheckError(t *testing.T) {
	expectedErr := errors.New("database error")
	fakeRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeRepo,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_Send_CreateError(t *testing.T) {
	expectedErr := errors.New("failed to save message")

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
			return expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_Send_Success(t *testing.T) {
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
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

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			return nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		publisher:     fakePublisher,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if createdMessage == nil {
		t.Fatalf("expected message to be created")
	}

	if createdMessage.ConversationID != 1 {
		t.Fatalf("expected conversationId 1, got %d", createdMessage.ConversationID)
	}

	if createdMessage.SenderID != 1 {
		t.Fatalf("expected sender ID 1, got %d", createdMessage.SenderID)
	}

	if createdMessage.Content != "Hello, world" {
		t.Fatalf("expected content %q, got %q", "Hello, world", createdMessage.Content)
	}
}

func TestMessageService_Send_BroadcastsMessage(t *testing.T) {
	var publishedEvent *websocket.BroadcastEvent

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			publishedEvent = &event
			return nil
		},
	}

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
		publisher:     fakePublisher,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedRecipients := []int64{1, 2}
	if len(publishedEvent.RecipientIDs) != len(expectedRecipients) {
		t.Fatalf("expected %d deliveries, got %d", len(expectedRecipients), len(publishedEvent.RecipientIDs))
	}

	for i, expectedID := range expectedRecipients {
		if publishedEvent.RecipientIDs[i] != expectedID {
			t.Fatalf("expected recipient index %d to be %d, got %d", i, expectedID, publishedEvent.RecipientIDs[i])
		}
	}

	if publishedEvent.Event.Type != websocket.EventMessage {
		t.Fatalf("expected event type %q, got %q", websocket.EventMessage, publishedEvent.Event.Type)
	}

	payloadBytes, err := json.Marshal(publishedEvent.Event.Payload)
	if err != nil {
		t.Fatalf("failed to re-marshal payload: %v", err)
	}

	var messageEvent MessageEvent
	if err := json.Unmarshal(payloadBytes, &messageEvent); err != nil {
		t.Fatalf("failed to decode message payload: %v", err)
	}

	if messageEvent.ID == 0 {
		t.Fatal("expected message ID to be set")
	}

	if messageEvent.ConversationID != 1 {
		t.Fatalf("expected conversation ID 1, got %d", messageEvent.ConversationID)
	}

	if messageEvent.SenderID != 1 {
		t.Fatalf("expected sender ID 1, got %d", messageEvent.SenderID)
	}

	if messageEvent.Content != "Hello, world" {
		t.Fatalf("expected content %q, got %q", "Hello, world", messageEvent.Content)
	}
}

func TestMessageService_Send_ParticipantsError(t *testing.T) {
	var publishedEvent *websocket.BroadcastEvent

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			publishedEvent = &event
			return nil
		},
	}

	expectedErr := errors.New("failed to fetch participants")

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
		publisher:     fakePublisher,
	}

	_, err := service.Send(context.Background(), 1, 1, "Hello, world")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if publishedEvent != nil {
		t.Fatal("expected no event to be published")
	}
}

func TestMessageService_MarkAsRead_NotParticipant(t *testing.T) {
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
	}

	_, err := service.MarkAsRead(context.Background(), 1, 1, 1)

	if !errors.Is(err, domain.ErrNotConversationParticipant) {
		t.Fatalf("expected ErrNotConversationParticipant, got %v", err)
	}
}

func TestMessageService_MarkAsRead_ParticipantCheckError(t *testing.T) {
	expectedErr := errors.New("database error")
	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
	}

	_, err := service.MarkAsRead(context.Background(), 1, 1, 1)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_MarkAsRead_MessageNotFound(t *testing.T) {
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

	_, err := service.MarkAsRead(context.Background(), 1, 1, 1)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_MarkAsRead_RepositoryError(t *testing.T) {
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

	_, err := service.MarkAsRead(context.Background(), 1, 1, 1)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_MarkAsRead_Success(t *testing.T) {
	expectedReadAt := time.Now()

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID int64, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
			return []int64{1, 2, 3}, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return expectedReadAt, nil
		},
	}

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			return nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
		publisher:     fakePublisher,
	}

	readAt, err := service.MarkAsRead(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if readAt.IsZero() {
		t.Fatal("expected readAt to be set")
	}

	if !readAt.Equal(expectedReadAt) {
		t.Fatalf("expected readAt %v, got %v", expectedReadAt, readAt)
	}
}

func TestMessageService_MarkAsRead_ParticipantsError(t *testing.T) {
	expectedErr := errors.New("failed to fetch participants")
	expectedReadAt := time.Now()

	var publishedEvent *websocket.BroadcastEvent

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			publishedEvent = &event
			return nil
		},
	}

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
		publisher:     fakePublisher,
	}

	_, err := service.MarkAsRead(context.Background(), 1, 1, 1)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if publishedEvent != nil {
		t.Fatal("expected no event to be published")
	}
}

func TestMessageService_MarkAsRead_BroadcastsReadReceipt(t *testing.T) {
	var publishedEvent *websocket.BroadcastEvent

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			publishedEvent = &event
			return nil
		},
	}

	expectedReadAt := time.Now()

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID int64, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
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
		publisher:     fakePublisher,
	}

	readAt, err := service.MarkAsRead(context.Background(), 1, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !readAt.Equal(expectedReadAt) {
		t.Fatalf("expected readAt %v, got %v", expectedReadAt, readAt)
	}

	if publishedEvent == nil {
		t.Fatal("expected event to be published")
	}

	expectedRecipients := []int64{2}
	if len(publishedEvent.RecipientIDs) != len(expectedRecipients) {
		t.Fatalf("expected %d recipients, got %d", len(expectedRecipients), len(publishedEvent.RecipientIDs))
	}

	for i, expectedID := range expectedRecipients {
		if publishedEvent.RecipientIDs[i] != expectedID {
			t.Fatalf("expected recipient index %d to be %d, got %d", i, expectedID, publishedEvent.RecipientIDs[i])
		}
	}

	if publishedEvent.Event.Type != websocket.EventReadReceipt {
		t.Fatalf("expected event type %q, got %q", websocket.EventReadReceipt, publishedEvent.Event.Type)
	}

	readReceipt, ok := publishedEvent.Event.Payload.(websocket.ReadReceiptEvent)
	if !ok {
		t.Fatalf("expected ReadReceiptEvent payload, got %T", publishedEvent.Event.Payload)
	}

	if readReceipt.MessageID != 1 {
		t.Fatalf("expected message ID 1, got %d", readReceipt.MessageID)
	}

	if readReceipt.ConversationID != 1 {
		t.Fatalf("expected conversation ID 1, got %d", readReceipt.ConversationID)
	}

	if readReceipt.UserID != 1 {
		t.Fatalf("expected user ID 1, got %d", readReceipt.UserID)
	}

	if !readReceipt.ReadAt.Equal(expectedReadAt) {
		t.Fatalf("expected readAt %v, got %v", expectedReadAt, readReceipt.ReadAt)
	}
}

func TestMessageService_MarkAsRead_BroadcastsToAllOtherParticipants(t *testing.T) {
	var publishedEvent *websocket.BroadcastEvent

	fakePublisher := &fakeEventPublisher{
		publishFunc: func(ctx context.Context, event websocket.BroadcastEvent) error {
			publishedEvent = &event
			return nil
		},
	}

	expectedTime := time.Now()

	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
		ParticipantsFunc: func(ctx context.Context, conversationID int64) ([]int64, error) {
			return []int64{1, 2, 3}, nil
		},
	}

	fakeMessageRepository := &fakeMessageRepository{
		MarkAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return expectedTime, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepository,
		messages:      fakeMessageRepository,
		publisher:     fakePublisher,
	}

	_, err := service.MarkAsRead(context.Background(), 24, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publishedEvent == nil {
		t.Fatal("expected event to be published")
	}

	expectedRecipients := []int64{2, 3}
	if len(publishedEvent.RecipientIDs) != len(expectedRecipients) {
		t.Fatalf("expected %d recipients, got %d", len(expectedRecipients), len(publishedEvent.RecipientIDs))
	}

	for i, expectedID := range expectedRecipients {
		if publishedEvent.RecipientIDs[i] != expectedID {
			t.Fatalf("expected recipient index %d to be %d, got %d", i, expectedID, publishedEvent.RecipientIDs[i])
		}
	}
}

func TestMessageService_ListForConversation_NotParticipant(t *testing.T) {
	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepository,
	}

	_, err := service.ListForConversation(context.Background(), 1, 1, 20, nil)

	if !errors.Is(err, domain.ErrNotConversationParticipant) {
		t.Fatalf("expected ErrNotConversationParticipant, got %v", err)
	}
}

func TestMessageService_ListForConversation_ParticipantCheckError(t *testing.T) {
	expectedErr := errors.New("database error")
	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return false, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepository,
	}

	_, err := service.ListForConversation(context.Background(), 1, 1, 10, nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_ListForConversation_RepositoryError(t *testing.T) {
	expectedErr := errors.New("failed to fetch messages")

	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepository := &fakeMessageRepository{
		ListForConversationFunc: func(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
			return nil, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepository,
		messages:      fakeMessageRepository,
	}

	_, err := service.ListForConversation(context.Background(), 1, 1, 10, nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_ReadReceipt_RepositoryError(t *testing.T) {
	expectedErr := errors.New("failed to fetch read receipts")

	fakeConversationRepository := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepository := &fakeMessageRepository{
		ListForConversationFunc: func(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
			return []*models.Message{}, nil
		},
		GetReadReceiptsFunc: func(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
			return nil, expectedErr
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepository,
		messages:      fakeMessageRepository,
	}

	_, err := service.ListForConversation(context.Background(), 1, 1, 10, nil)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestMessageService_ListForConversation_Success(t *testing.T) {
	message1 := &models.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "Hello",
	}

	message2 := &models.Message{
		ID:             2,
		ConversationID: 1,
		SenderID:       2,
		Content:        "Hi",
	}

	expectedMessages := []*models.Message{message1, message2}
	expectedReads := []*domain.MessageRead{
		{
			MessageID: 1,
			UserID:    2,
		},
	}

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID int64, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		ListForConversationFunc: func(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
			return expectedMessages, nil
		},
		GetReadReceiptsFunc: func(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
			return expectedReads, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	result, err := service.ListForConversation(context.Background(), 1, 1, 20, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}

	if result.Messages[0].ID != 1 {
		t.Fatalf("expected first message ID 1, got %d", result.Messages[0].ID)
	}

	if result.Messages[1].ID != 2 {
		t.Fatalf("expected second message ID 2, got %d", result.Messages[1].ID)
	}

	if len(result.Reads) != 1 {
		t.Fatalf("expected 1 read receipt, got %d", len(result.Reads))
	}

	if result.Reads[0].MessageID != 1 {
		t.Fatalf("expected read receipt for message 1, got %d", result.Reads[0].MessageID)
	}

	if result.HasMore {
		t.Fatal("expected HasMore to be false")
	}

	if result.NextCursor != nil {
		t.Fatal("expected NextCursor to be nil")
	}
}

func TestMessageService_ListForConversation_HasMore(t *testing.T) {
	// Arrange: fetch limit of 2, but repository returns 3 items (limit + 1 keyset pattern)
	limit := 2
	messages := []*models.Message{
		{ID: 103},
		{ID: 102},
		{ID: 101}, // Extra item signaling a next page
	}

	fakeConversationRepo := &fakeConversationRepository{
		IsParticipantFunc: func(ctx context.Context, conversationID int64, userID int64) (bool, error) {
			return true, nil
		},
	}

	fakeMessageRepo := &fakeMessageRepository{
		ListForConversationFunc: func(ctx context.Context, conversationID int64, before *int64, limit int) ([]*models.Message, error) {
			return messages, nil
		},
		GetReadReceiptsFunc: func(ctx context.Context, conversationID int64) ([]*domain.MessageRead, error) {
			return []*domain.MessageRead{}, nil
		},
	}

	service := &MessageService{
		conversations: fakeConversationRepo,
		messages:      fakeMessageRepo,
	}

	// Act
	result, err := service.ListForConversation(context.Background(), 1, 1, limit, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.HasMore {
		t.Fatal("expected HasMore to be true")
	}

	if len(result.Messages) != limit {
		t.Fatalf("expected returned messages length to match limit (%d), got %d", limit, len(result.Messages))
	}

	if result.NextCursor == nil {
		t.Fatal("expected NextCursor to be non-nil")
	}

	// Cursor should point to the last item in the truncated list (ID 102)
	expectedCursor := int64(102)
	if *result.NextCursor != expectedCursor {
		t.Fatalf("expected NextCursor %d, got %d", expectedCursor, *result.NextCursor)
	}
}
