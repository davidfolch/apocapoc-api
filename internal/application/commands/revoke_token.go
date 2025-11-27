package commands

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type RevokeTokenCommand struct {
	RefreshToken string
}

type RevokeTokenHandler struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewRevokeTokenHandler(refreshTokenRepo repositories.RefreshTokenRepository) *RevokeTokenHandler {
	return &RevokeTokenHandler{
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (h *RevokeTokenHandler) Handle(ctx context.Context, cmd RevokeTokenCommand) error {
	if cmd.RefreshToken == "" {
		return errors.ErrInvalidInput
	}

	return h.refreshTokenRepo.RevokeByToken(ctx, cmd.RefreshToken)
}
