package commands

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type RevokeAllTokensCommand struct {
	UserID string
}

type RevokeAllTokensHandler struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewRevokeAllTokensHandler(refreshTokenRepo repositories.RefreshTokenRepository) *RevokeAllTokensHandler {
	return &RevokeAllTokensHandler{
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (h *RevokeAllTokensHandler) Handle(ctx context.Context, cmd RevokeAllTokensCommand) error {
	if cmd.UserID == "" {
		return errors.ErrInvalidInput
	}

	return h.refreshTokenRepo.RevokeAllByUserID(ctx, cmd.UserID)
}
