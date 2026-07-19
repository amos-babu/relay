package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"relay/internal/domain"
	"relay/internal/response"
	"relay/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err := response.JSON(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid request body",
		})
		if err != nil {
			log.Printf("failed to encode response: %v", err)
		}

		return
	}

	user, err := h.service.Register(
		r.Context(),
		req.Name,
		req.Email,
		req.Password,
	)

	if err != nil {
		if err := response.JSON(
			w,
			http.StatusInternalServerError,
			response.ErrorResponse{
				Error: "internal server error",
			}); err != nil {
			log.Printf("failed to register: %v", err)
		}
		return
	}

	resp := RegisterResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	if err := response.JSON(
		w,
		http.StatusCreated,
		resp,
	); err != nil {
		log.Printf("failed to encode response: %v", err)
	}

	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			if encodeErr := response.JSON(
				w,
				http.StatusConflict,
				response.ErrorResponse{
					Error: "email already exists",
				},
			); encodeErr != nil {
				log.Printf("failed to encode response: %v", encodeErr)
			}
			return
		}

		if encodeErr := response.JSON(
			w,
			http.StatusInternalServerError,
			response.ErrorResponse{
				Error: "internal server error",
			},
		); encodeErr != nil {
			log.Printf("failed to encode response: %v", encodeErr)
		}

		return
	}
}
