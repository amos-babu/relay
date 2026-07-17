package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErroResponse struct {
	Error string `json:"error"`
}

func Rocovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				if err := json.NewEncoder(w).Encode(ErroResponse{
					Error: "Internal Server Error",
				}); err != nil {
					log.Printf("failed to encode error response: %v", err)
				}
			}
		}()

		next.ServeHTTP(w, r)

	})
}
