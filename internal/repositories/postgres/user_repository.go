package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/models"
	"relay/internal/repositories"
)

var ErrNotImplemented = errors.New("not implemented")

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
	const query = `INSERT INTO users (
	name, 
	email, 
	password_hash, 
	created_at, 
	updated_at) VALUES (
	$1, $2, $3, $4, $5) 
	RETURNING id, created_at, updated_at;`

	return r.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const query = `SELECT
			id,
			name,
			email,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE id = $1;`

	user := &models.User{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, ErrNotImplemented
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return ErrNotImplemented
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return ErrNotImplemented
}
