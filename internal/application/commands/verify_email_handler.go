package commands

import (
	"context"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type VerifyEmailCommand struct {
	Token string
}

type VerifyEmailHandler struct {
	userRepo repositories.UserRepository
}

func NewVerifyEmailHandler(userRepo repositories.UserRepository) *VerifyEmailHandler {
	return &VerifyEmailHandler{
		userRepo: userRepo,
	}
}

func (h *VerifyEmailHandler) Handle(ctx context.Context, cmd VerifyEmailCommand) error {
	if cmd.Token == "" {
		return errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByVerificationToken(ctx, cmd.Token)
	if err != nil {
		return errors.ErrInvalidInput
	}

	if user.EmailVerified {
		return errors.ErrAlreadyExists
	}

	if user.EmailVerificationExpiry == nil || user.EmailVerificationExpiry.Before(time.Now()) {
		return errors.ErrInvalidInput
	}

	user.EmailVerified = true
	user.EmailVerificationToken = nil
	user.EmailVerificationExpiry = nil
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}
