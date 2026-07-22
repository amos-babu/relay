package postgres

import (
	"context"
	"relay/internal/models"
)

func Create(ctx context.Context, token *models.RefreshToken) error {
	return nil
}
func GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return nil, nil
}
func Revoke(ctx context.Context, id int64) error {
	return nil
}
