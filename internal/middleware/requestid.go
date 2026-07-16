package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()

		ctx := context.WithValue(
			r.Context(),
			requestIDKey,
			id,
		)

		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)

	})
}
