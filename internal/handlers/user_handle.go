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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if encodeErr := response.JSON(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid request body",
		}); encodeErr != nil {
			log.Printf("failed to encode response: %v", encodeErr)
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
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			if encodeErr := response.JSON(w, http.StatusConflict, response.ErrorResponse{
				Error: "email already exists",
			}); encodeErr != nil {
				log.Printf("failed to encode response: %v", encodeErr)
			}
			return
		}

		if encodeErr := response.JSON(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "internal server error",
		}); encodeErr != nil {
			log.Printf("failed to encode response: %v", encodeErr)
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

}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if encodeErr := response.JSON(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "invalid request body",
		}); encodeErr != nil {
			log.Printf("failed to encode response: %v", encodeErr)
		}
		return
	}

	result, err := h.service.Login(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			if encodeErr := response.JSON(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "invalid email or password",
			}); encodeErr != nil {
				log.Printf("failed to encode response: %v", encodeErr)
			}
			return
		}

		if encodeErr := response.JSON(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "internal server error",
		}); encodeErr != nil {
			log.Printf("failed to encode response: %v", encodeErr)
		}
		return
	}

	resp := LoginResponse{
		Token: result.Token,
		User: UserResponse{
			ID:    result.User.ID,
			Name:  result.User.Name,
			Email: result.User.Email,
		},
	}

	if err := response.JSON(
		w,
		http.StatusOK,
		resp,
	); err != nil {
		log.Printf("failed to encode response: %v", err)
	}

}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {

}
