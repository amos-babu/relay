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
	"time"
)

type ConversationHandle struct {
	service *services.ConversationService
}

type CreateConversationRequest struct {
	RecipientID int64 `json:"recipient_id"`
}

type ConversationResponse struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewConversationHandle(service *services.ConversationService) *ConversationHandle {
	return &ConversationHandle{
		service: service,
	}
}

func (h *ConversationHandle) Create(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	creatorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "unauthorized")
		return
	}

	//Decode request
	var req CreateConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Call service
	conversation, err := h.service.Create(
		r.Context(),
		creatorID,
		req.RecipientID,
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotMessageYourself):
			response.BadRequest(w, "cannot message yourself")
		case errors.Is(err, domain.ErrUserNotFound):
			response.NotFound(w, "recipient not found")
		default:
			response.InternalServerError(w)
		}
		return
	}

	resp := ConversationResponse{
		ID:        conversation.ID,
		CreatedAt: conversation.CreatedAt,
	}

	if err := response.JSON(
		w,
		http.StatusCreated,
		resp,
	); err != nil {
		log.Printf("failed to encode response: %v", err)
	}

}

func (h *ConversationHandle) ListForUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "unauthorized")
		return
	}

	conversations, err := h.service.ListForUser(
		r.Context(),
		userID,
	)
	if err != nil {
		response.InternalServerError(w)
		return
	}

	resp := make([]ConversationResponse, len(conversations))
	for i, conversation := range conversations {
		resp[i] = ConversationResponse{
			ID:        conversation.ID,
			CreatedAt: conversation.CreatedAt,
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
