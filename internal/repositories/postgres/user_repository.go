package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"

	"github.com/jackc/pgx/v5/pgconn"
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
	const query = `
	INSERT INTO users (
		name, 
		email, 
		password_hash
	) 
	VALUES (
		$1, 
		$2, 
		$3) 
	RETURNING 
		id, 
		created_at, 
		updated_at;
	`

	err := r.db.QueryRowContext(
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

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyExists
		}

		return err
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
	SELECT
		id,
		name,
		email,
		password_hash,
		created_at,
		updated_at
	FROM users
	WHERE email = $1;
	`
	user := &models.User{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const query = `
	SELECT
		id,
		name,
		email,
		password_hash,
		created_at,
		updated_at
	FROM users
	WHERE id = $1;
	`

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

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return ErrNotImplemented
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return ErrNotImplemented
}
