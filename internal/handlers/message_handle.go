package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"relay/internal/domain"
	"relay/internal/middleware"
	"relay/internal/response"
	"relay/internal/services"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type MessageHandle struct {
	service *services.MessageService
}

func NewMessageHandle(service *services.MessageService) *MessageHandle {
	return &MessageHandle{
		service: service,
	}
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderID       int64     `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func (h *MessageHandle) Send(w http.ResponseWriter, r *http.Request) {
	conversationIDStr := chi.URLParam(r, "conversationID")

	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid conversation id")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req SendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	message, err := h.service.Send(
		r.Context(),
		conversationID,
		userID,
		req.Content,
	)
	//

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmptyMessage):
			response.BadRequest(w, "empty message")
		case errors.Is(err, domain.ErrNotConversationParticipant):
			response.Forbidden(w, "user is not a participant in this conversation")
		default:
			response.InternalServerError(w)
		}
		return
	}

	resp := MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}

	if err := response.JSON(
		w,
		http.StatusCreated,
		resp,
	); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *MessageHandle) ListForConversation(w http.ResponseWriter, r *http.Request) {
	conversationIDStr := chi.URLParam(r, "conversationID")

	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid conversation id")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "unauthorized")
		return
	}

	messages, err := h.service.ListForConversation(r.Context(), conversationID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotConversationParticipant) {
			response.Forbidden(w, "user is not a participant in this conversation")
			return
		}

		response.InternalServerError(w)
		return
	}

	resp := make([]MessageResponse, len(messages))
	for i, message := range messages {
		resp[i] = MessageResponse{
			ID:             message.ID,
			ConversationID: message.ConversationID,
			SenderID:       message.SenderID,
			Content:        message.Content,
			CreatedAt:      message.CreatedAt,
		}
	}

	if err := response.JSON(
		w,
		http.StatusOK,
		resp,
	); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
