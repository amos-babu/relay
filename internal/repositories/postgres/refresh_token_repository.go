package postgres

import (
	"context"
	"database/sql"
	"errors"
	"relay/internal/domain"
	"relay/internal/models"
	"relay/internal/repositories"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

var _ repositories.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	const query = `
	INSERT INTO refresh_tokens (
		user_id,
		token_hash,
		expires_at
	) 
	VALUES (
		$1, 
		$2, 
		$3
	) RETURNING 
		id, 
		created_at;
	`
	if err := r.db.QueryRowContext(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(
		&token.ID,
		&token.CreatedAt,
	); err != nil {
		return err
	}

	return nil
}
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	const query = `
	SELECT 
		id,
		user_id,
		token_hash,
		expires_at,
		created_at,
		revoked_at
	FROM refresh_tokens 
	WHERE token_hash = $1;`

	refreshToken := &models.RefreshToken{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		hash,
	).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
		&refreshToken.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	return refreshToken, nil
}
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int64) error {
	const query = `
	UPDATE refresh_tokens
	SET revoked_at = NOW()
	WHERE id = $1;`
	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrRefreshTokenNotFound
	}
	return nil
}
