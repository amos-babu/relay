package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/models"
	"relay/internal/repositories"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

var _ repositories.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return errors.New("not implemented")
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return errors.New("not implemented")
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}
