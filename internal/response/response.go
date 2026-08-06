package response

import (
	"encoding/json"
	"log"
	"net/http"
	"relay/internal/validation"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Errors map[string]string `json:"errors"`
}

// type ValidationResponse struct {
// 	Error  string            `json:"error"`
// 	Errors map[string]string `json:"errors"`
// }

func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
	if err := JSON(w, status, ErrorResponse{
		Error: message,
	}); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message)
}

func UnprocessableEntity(w http.ResponseWriter, vErr *validation.ValidationError) {
	if err := JSON(
		w,
		http.StatusUnprocessableEntity,
		ValidationErrorResponse{
			Error:  "validation failed",
			Errors: vErr.Errors,
		}, //422
	); err != nil {
		log.Printf("failed to encode response: %v", err)

	}
}

func InternalServerError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}
