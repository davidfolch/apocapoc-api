package queries

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type RefreshTokenQuery struct {
	RefreshToken string
}

type RefreshTokenResult struct {
	UserID string
	Email  string
}

type RefreshTokenHandler struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	userRepo         repositories.UserRepository
}

func NewRefreshTokenHandler(
	refreshTokenRepo repositories.RefreshTokenRepository,
	userRepo repositories.UserRepository,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, query RefreshTokenQuery) (*RefreshTokenResult, error) {
	if query.RefreshToken == "" {
		return nil, errors.ErrInvalidInput
	}

	refreshToken, err := h.refreshTokenRepo.FindByToken(ctx, query.RefreshToken)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if !refreshToken.IsValid() {
		return nil, errors.ErrNotFound
	}

	user, err := h.userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	return &RefreshTokenResult{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func CreateRefreshToken(userID string, expiryDuration time.Duration) (*entities.RefreshToken, error) {
	token, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(expiryDuration)
	return entities.NewRefreshToken(userID, token, expiresAt), nil
}
