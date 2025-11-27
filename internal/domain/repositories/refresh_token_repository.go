package repositories

import (
	"context"

	"apocapoc-api/internal/domain/entities"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
