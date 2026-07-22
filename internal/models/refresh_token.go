package models

import "time"

type RefreshToken struct {
	ID int64

	UserID int64

	TokenHash string

	ExpiresAt time.Time

	CreatedAt time.Time

	RevokedAt *time.Time
}
