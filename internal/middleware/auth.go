package middleware

import (
	"context"
	"log"
	"net/http"
	"relay/internal/response"
	"relay/internal/token"
	"strings"
)

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

func Auth(tokens *token.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			//check if the header is missing
			if authHeader == "" {
				if encodeErr := response.JSON(
					w,
					http.StatusUnauthorized,
					response.ErrorResponse{
						Error: "authorization header is missing",
					},
				); encodeErr != nil {
					log.Printf("failed to encode response: %v", encodeErr)
				}
				return
			}
			//check if the header starts with bearer
			const bearer = "Bearer "
			if !strings.HasPrefix(authHeader, bearer) {
				if encodeErr := response.JSON(
					w,
					http.StatusUnauthorized,
					response.ErrorResponse{
						Error: "invalid authorization header",
					},
				); encodeErr != nil {
					log.Printf("failed to encode response: %v", encodeErr)
				}
				return
			}

			//Extract the token
			tokenString := strings.TrimPrefix(authHeader, bearer)

			//Validate
			userID, err := tokens.Validate(tokenString)

			//If validation fails
			if err != nil {
				if encodeErr := response.JSON(
					w,
					http.StatusUnauthorized,
					response.ErrorResponse{
						Error: "invalid or expired token",
					},
				); encodeErr != nil {
					log.Printf("failed to encode response: %v", encodeErr)
				}
				return
			}

			// Put the user ID in a new context
			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				userID,
			)

			r = r.WithContext(ctx)

			// Next handler now has access to the context
			next.ServeHTTP(w, r)
		})
	}
}
