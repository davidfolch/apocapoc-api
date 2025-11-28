package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type ResendVerificationEmailCommand struct {
	Email string
}

type ResendVerificationEmailHandler struct {
	userRepo     repositories.UserRepository
	emailService services.EmailService
	appURL       string
}

func NewResendVerificationEmailHandler(
	userRepo repositories.UserRepository,
	emailService services.EmailService,
	appURL string,
) *ResendVerificationEmailHandler {
	return &ResendVerificationEmailHandler{
		userRepo:     userRepo,
		emailService: emailService,
		appURL:       appURL,
	}
}

func (h *ResendVerificationEmailHandler) Handle(ctx context.Context, cmd ResendVerificationEmailCommand) error {
	if cmd.Email == "" {
		return errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return errors.ErrNotFound
	}

	if user.EmailVerified {
		return errors.ErrAlreadyExists
	}

	token, err := generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", h.appURL, token)

	emailBody := fmt.Sprintf(`
		<h2>Verify your email address</h2>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
		<p>If you didn't create an account, you can safely ignore this email.</p>
	`, verificationLink)

	message := services.EmailMessage{
		To:      user.Email,
		Subject: "Verify your email address",
		Body:    emailBody,
		IsHTML:  true,
	}

	if err := h.emailService.Send(message); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
