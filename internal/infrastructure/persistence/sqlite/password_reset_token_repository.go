package sqlite

import (
	"context"
	"database/sql"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type PasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(db *sql.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *entities.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at, used_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
		token.UsedAt,
	)

	return err
}

func (r *PasswordResetTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*entities.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, used_at
		FROM password_reset_tokens
		WHERE token = ?
	`

	token := &entities.PasswordResetToken{}
	err := r.db.QueryRowContext(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *PasswordResetTokenRepository) Update(ctx context.Context, token *entities.PasswordResetToken) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, token.UsedAt, token.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
