package middleware

import (
	"net/http"
	"relay/internal/token"
)

func Auth(tokens *token.Service) func(http.Handler) http.Handler {

}
