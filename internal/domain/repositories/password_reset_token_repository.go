package repositories

import (
	"context"

	"apocapoc-api/internal/domain/entities"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *entities.PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*entities.PasswordResetToken, error)
	Update(ctx context.Context, token *entities.PasswordResetToken) error
	DeleteExpired(ctx context.Context) error
}
