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

				json.NewEncoder(w).Encode(ErroResponse{
					Error: "Internal Server Error",
				})
			}
		}()

		next.ServeHTTP(w, r)

	})
}
