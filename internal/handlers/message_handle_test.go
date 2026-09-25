package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"relay/internal/domain"
	"relay/internal/middleware"
	"relay/internal/models"
	"relay/internal/services"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type fakeMessageService struct {
	sendFunc func(
		ctx context.Context,
		conversationID int64,
		senderID int64,
		content string,
	) (*models.Message, error)

	listFunc func(
		ctx context.Context,
		conversationID int64,
		userID int64,
		limit int,
		before *int64,
	) (*services.ConversationMessages, error)

	markAsReadFunc func(
		ctx context.Context,
		messageID int64,
		conversationID int64,
		userID int64,
	) (time.Time, error)
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
	return f.listFunc(ctx, conversationID, userID, limit, before)
}

func (f *fakeMessageService) MarkAsRead(
	ctx context.Context,
	messageID int64,
	conversationID int64,
	userID int64,
) (time.Time, error) {
	return f.markAsReadFunc(ctx, messageID, conversationID, userID)
}

// SEND TESTS
func TestMessageHandler_Send_InvalidConversationID(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/not-a-number/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_InvalidUserAuthentication(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_InvalidJSONBody(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":}`),
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_EmptyMessage(t *testing.T) {
	fakeService := &fakeMessageService{
		sendFunc: func(ctx context.Context, conversationID, senderID int64, content string) (*models.Message, error) {
			return nil, domain.ErrEmptyMessage
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":""}`),
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_NotParticipant(t *testing.T) {
	fakeService := &fakeMessageService{
		sendFunc: func(ctx context.Context, conversationID, senderID int64, content string) (*models.Message, error) {
			return nil, domain.ErrNotConversationParticipant
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_InsternalServerError(t *testing.T) {
	fakeService := &fakeMessageService{
		sendFunc: func(ctx context.Context, conversationID, senderID int64, content string) (*models.Message, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestMessageHandler_Send_Successful(t *testing.T) {
	var gotConversationID int64
	var gotSenderID int64
	var gotContent string

	fakeService := &fakeMessageService{

		sendFunc: func(ctx context.Context, conversationID, senderID int64, content string) (*models.Message, error) {
			gotConversationID = conversationID
			gotSenderID = senderID
			gotContent = content
			return &models.Message{
				ID:             100,
				ConversationID: conversationID,
				SenderID:       senderID,
				Content:        content,
			}, nil
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.Send,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if gotConversationID != 123 {
		t.Errorf(
			"expected conversation ID 123, got %d",
			gotConversationID,
		)
	}

	if gotSenderID != 42 {
		t.Errorf(
			"expected sender ID 42, got %d",
			gotSenderID,
		)
	}

	if gotContent != "hello" {
		t.Errorf(
			"expected content %q, got %q",
			"hello",
			gotContent,
		)
	}

	var response MessageResponse

	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != 100 {
		t.Fatalf("expected response ID 100, got %d", response.ID)
	}

	if response.ConversationID != 123 {
		t.Fatalf(
			"expected response conversation ID 123, got %d",
			response.ConversationID,
		)
	}

	if response.SenderID != 42 {
		t.Fatalf(
			"expected response sender ID 42, got %d",
			response.SenderID,
		)
	}

	if response.Content != "hello" {
		t.Fatalf(
			"expected response content hello, got %s",
			response.Content,
		)
	}
}

// LIST FOR CONVESATION TESTS
func TestMessageHandler_ListforConversation_InvalidConversationID(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/not-a-number/messages",
		strings.NewReader(`{"content":"hello"}`),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_ListforConversation_InvalidLimit(t *testing.T) {
	fakeService := &fakeMessageService{
		listFunc: func(ctx context.Context, conversationID, userID int64, limit int, before *int64) (*services.ConversationMessages, error) {
			t.Fatal("service should not be called")
			return nil, nil
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages?limit=abc",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_ZeroLimit(t *testing.T) {
	fakeService := &fakeMessageService{
		listFunc: func(ctx context.Context, conversationID, userID int64, limit int, before *int64) (*services.ConversationMessages, error) {
			t.Fatal("service should not be called")
			return nil, nil
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages?limit=0",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_LimitCappedAt100(t *testing.T) {
	var gotLimit int

	fakeService := &fakeMessageService{
		listFunc: func(ctx context.Context, conversationID, userID int64, limit int, before *int64) (*services.ConversationMessages, error) {
			gotLimit = limit

			return &services.ConversationMessages{
				Messages: []*models.Message{},
				Reads:    []*domain.MessageRead{},
				HasMore:  false,
			}, nil
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages?limit=150",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	if gotLimit != 100 {
		t.Fatalf(
			"expected service to receive limit 100, got %d",
			gotLimit,
		)
	}
}

func TestMessageHandler_ListForConversation_InvalidBefore(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/123/messages?before=abc",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_Unauthorized(t *testing.T) {
	fakeService := &fakeMessageService{}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/123/messages",
		nil,
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status 401, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_NotParticipant(t *testing.T) {
	fakeService := &fakeMessageService{
		listFunc: func(
			ctx context.Context,
			conversationID int64,
			userID int64,
			limit int,
			before *int64,
		) (*services.ConversationMessages, error) {
			return nil, domain.ErrNotConversationParticipant
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/123/messages",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_InternalServerError(t *testing.T) {
	fakeService := &fakeMessageService{
		listFunc: func(
			ctx context.Context,
			conversationID int64,
			userID int64,
			limit int,
			before *int64,
		) (*services.ConversationMessages, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/123/messages",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_ListForConversation_Success(t *testing.T) {
	// before := int64(50)

	fakeService := &fakeMessageService{
		listFunc: func(
			ctx context.Context,
			conversationID int64,
			userID int64,
			limit int,
			gotBefore *int64,
		) (*services.ConversationMessages, error) {

			if conversationID != 123 {
				t.Fatalf("expected conversation ID 123, got %d", conversationID)
			}

			if userID != 42 {
				t.Fatalf("expected user ID 42, got %d", userID)
			}

			if limit != 20 {
				t.Fatalf("expected limit 20, got %d", limit)
			}

			if gotBefore == nil {
				t.Fatal("expected before cursor")
			}

			if *gotBefore != 50 {
				t.Fatalf("expected before cursor 50, got %d", *gotBefore)
			}

			return &services.ConversationMessages{
				Messages: []*models.Message{
					{
						ID:             100,
						ConversationID: 123,
						SenderID:       42,
						Content:        "Hello",
					},
					{
						ID:             99,
						ConversationID: 123,
						SenderID:       7,
						Content:        "Hi there",
					},
				},
				Reads: []*domain.MessageRead{
					{
						MessageID: 100,
						UserID:    7,
					},
				},
				NextCursor: func() *int64 {
					v := int64(99)
					return &v
				}(),
				HasMore: true,
			}, nil
		},
	}

	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/conversations/123/messages?before=50",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get(
		"/conversations/{conversationID}/messages",
		handler.ListForConversation,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	var response PaginatedMessageResponse

	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Messages) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(response.Messages),
		)
	}

	if response.Messages[0].ID != 100 {
		t.Fatalf(
			"expected first message ID 100, got %d",
			response.Messages[0].ID,
		)
	}

	if response.Messages[0].Content != "Hello" {
		t.Fatalf(
			"expected first message content Hello, got %s",
			response.Messages[0].Content,
		)
	}

	if response.Messages[1].ID != 99 {
		t.Fatalf(
			"expected second message ID 99, got %d",
			response.Messages[1].ID,
		)
	}

	if response.Messages[1].Content != "Hi there" {
		t.Fatalf(
			"expected second message content Hi there, got %s",
			response.Messages[1].Content,
		)
	}

	if response.NextCursor == nil {
		t.Fatal("expected next cursor")
	}

	if *response.NextCursor != 99 {
		t.Fatalf(
			"expected next cursor 99, got %d",
			*response.NextCursor,
		)
	}

	if !response.HasMore {
		t.Fatal("expected has_more to be true")
	}
}

// MARK AS READ TESTS
func TestMessageHandler_MarkAsRead_InvalidConversationID(t *testing.T) {
	fakeService := &fakeMessageService{}
	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/not-a-number/messages/100/read",
		nil,
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages/{messageID}/read",
		handler.MarkAsRead,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_MarkAsRead_InvalidMessageID(t *testing.T) {
	fakeService := &fakeMessageService{}
	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages/not-a-number/read",
		nil,
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages/{messageID}/read",
		handler.MarkAsRead,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_MarkAsRead_UnathorizedUser(t *testing.T) {
	fakeService := &fakeMessageService{}
	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages/100/read",
		nil,
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages/{messageID}/read",
		handler.MarkAsRead,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status 401, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_MarkAsRead_UserNotParticipant(t *testing.T) {
	fakeService := &fakeMessageService{
		markAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return time.Time{}, domain.ErrNotConversationParticipant
		},
	}
	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages/100/read",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages/{messageID}/read",
		handler.MarkAsRead,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			rec.Code,
		)
	}
}

func TestMessageHandler_MarkAsRead_InternalServerError(t *testing.T) {
	fakeService := &fakeMessageService{
		markAsReadFunc: func(ctx context.Context, messageID, conversationID, userID int64) (time.Time, error) {
			return time.Time{}, errors.New("database connection failed")
		},
	}
	handler := NewMessageHandler(fakeService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/conversations/123/messages/100/read",
		nil,
	)

	req = req.WithContext(
		middleware.ContextWithUserID(req.Context(), 42),
	)

	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post(
		"/conversations/{conversationID}/messages/{messageID}/read",
		handler.MarkAsRead,
	)

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			rec.Code,
		)
	}
}
