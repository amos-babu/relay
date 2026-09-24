package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
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
