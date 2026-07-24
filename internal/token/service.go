package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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

func (s *Service) GenerateAccessToken(userID int64) (string, error) {
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

func (s *Service) GenerateRefreshToken() (string, error) {
	//Generate secure 32 random bytes and then encode it.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) ValidateAccessToken(tokenStr string) (int64, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return s.secret, nil
		},
	)

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, err := strconv.ParseInt(
		claims.Subject,
		10,
		64,
	)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
