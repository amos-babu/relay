package repositories

import (
	"context"
	"relay/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error

	GetByEmail(ctx context.Context, email string) (*models.User, error)

	GetByID(ctx context.Context, id int64) (*models.User, error)

	Update(ctx context.Context, user *models.User) error

	Delete(ctx context.Context, id int64) error
}
