package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type RequestPasswordResetCommand struct {
	Email string
}

type RequestPasswordResetHandler struct {
	userRepo               repositories.UserRepository
	passwordResetTokenRepo repositories.PasswordResetTokenRepository
	emailService           services.EmailService
	appURL                 string
}

func NewRequestPasswordResetHandler(
	userRepo repositories.UserRepository,
	passwordResetTokenRepo repositories.PasswordResetTokenRepository,
	emailService services.EmailService,
	appURL string,
) *RequestPasswordResetHandler {
	return &RequestPasswordResetHandler{
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		emailService:           emailService,
		appURL:                 appURL,
	}
}

func (h *RequestPasswordResetHandler) Handle(ctx context.Context, cmd RequestPasswordResetCommand) error {
	if cmd.Email == "" {
		return errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return errors.ErrNotFound
	}

	if !user.EmailVerified {
		return errors.ErrEmailNotVerified
	}

	tokenStr, err := generatePasswordResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	resetToken := entities.NewPasswordResetToken(user.ID, tokenStr, expiresAt)

	if err := h.passwordResetTokenRepo.Create(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", h.appURL, tokenStr)

	emailBody := fmt.Sprintf(`
		<h2>Password Reset Request</h2>
		<p>You requested to reset your password. Click the link below to reset it:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link will expire in 1 hour.</p>
		<p>If you didn't request this, you can safely ignore this email.</p>
	`, resetLink)

	message := services.EmailMessage{
		To:      user.Email,
		Subject: "Password Reset Request",
		Body:    emailBody,
		IsHTML:  true,
	}

	if err := h.emailService.Send(message); err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

func generatePasswordResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
