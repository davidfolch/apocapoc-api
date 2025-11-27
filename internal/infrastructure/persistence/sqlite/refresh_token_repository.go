package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"

	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	token.ID = uuid.New().String()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
		token.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token = ?
	`

	var rt entities.RefreshToken
	var revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&revokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	if revokedAt.Valid {
		rt.RevokedAt = &revokedAt.Time
	}

	return &rt, nil
}

func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*entities.RefreshToken
	for rows.Next() {
		var rt entities.RefreshToken
		var revokedAt sql.NullTime

		err := rows.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.Token,
			&rt.ExpiresAt,
			&rt.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}

		if revokedAt.Valid {
			rt.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, &rt)
	}

	return tokens, nil
}

func (r *RefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE token = ? AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE user_id = ? AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}
