package commands

import (
	"context"
	"fmt"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
	"apocapoc-api/internal/shared/validation"
)

type ResetPasswordCommand struct {
	Token       string
	NewPassword string
}

type ResetPasswordHandler struct {
	userRepo               repositories.UserRepository
	passwordResetTokenRepo repositories.PasswordResetTokenRepository
	passwordHasher         services.PasswordHasher
}

func NewResetPasswordHandler(
	userRepo repositories.UserRepository,
	passwordResetTokenRepo repositories.PasswordResetTokenRepository,
	passwordHasher services.PasswordHasher,
) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		passwordHasher:         passwordHasher,
	}
}

func (h *ResetPasswordHandler) Handle(ctx context.Context, cmd ResetPasswordCommand) error {
	if cmd.Token == "" || cmd.NewPassword == "" {
		return errors.ErrInvalidInput
	}

	if err := validation.ValidatePassword(cmd.NewPassword); err != nil {
		return errors.ErrInvalidInput
	}

	resetToken, err := h.passwordResetTokenRepo.FindByToken(ctx, cmd.Token)
	if err != nil {
		return errors.ErrInvalidInput
	}

	if resetToken.IsExpired() {
		return errors.ErrInvalidInput
	}

	if resetToken.IsUsed() {
		return errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByID(ctx, resetToken.UserID)
	if err != nil {
		return errors.ErrNotFound
	}

	hashedPassword, err := h.passwordHasher.Hash(cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword

	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	resetToken.MarkAsUsed()
	if err := h.passwordResetTokenRepo.Update(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}
