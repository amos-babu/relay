package token

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret []byte
}

func NewService(secret string) *Service {
	return &Service{
		secret: []byte(secret),
	}
}

func (s *Service) Generate(userID int64) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject: strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(
			time.Now().Add(24 * time.Hour),
		),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secret)
}

func (s *Service) Validate(token string) (int64, error) {
	panic("not implemented")
}
