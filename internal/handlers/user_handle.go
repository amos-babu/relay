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
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

const refreshCookieName = "refresh_token" //__Host-refresh_token

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")

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
			response.Error(w, http.StatusConflict, "email already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
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
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Login(
		r.Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setRefreshCookie(w, result.RefreshToken)

	resp := LoginResponse{
		AccessToken: result.AccessToken,
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
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.Profile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	if encodeErr := response.JSON(
		w,
		http.StatusOK,
		resp,
	); encodeErr != nil {
		log.Printf("failed to encode response: %v", encodeErr)
	}
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	log.Printf("Cookie header: %q", r.Header.Get("Cookie"))
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.service.Refresh(
		r.Context(),
		cookie.Value,
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			h.clearRefreshCookie(w)
			response.Error(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal server error")

		return

	}

	h.setRefreshCookie(w, result.RefreshToken)

	if encodeErr := response.JSON(
		w,
		http.StatusOK,
		RefreshResponse{
			AccessToken: result.AccessToken,
		},
	); encodeErr != nil {
		log.Printf("failed to encode response: %v", encodeErr)
	}

}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		h.clearRefreshCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = h.service.Logout(
		r.Context(),
		cookie.Value,
	)

	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Helper function to set Cookie
func (h *UserHandler) setRefreshCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //Change to true when in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(services.RefreshTokenTTL),
		MaxAge:   int(services.RefreshTokenTTL.Seconds()),
	})
}

// Helper function to clear Cookie
func (h *UserHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
