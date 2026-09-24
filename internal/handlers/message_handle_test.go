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
