package commands

import (
	"context"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type VerifyEmailCommand struct {
	Token string
}

type VerifyEmailHandler struct {
	userRepo         repositories.UserRepository
	emailService     services.EmailService
	sendWelcomeEmail bool
}

func NewVerifyEmailHandler(
	userRepo repositories.UserRepository,
	emailService services.EmailService,
	sendWelcomeEmail bool,
) *VerifyEmailHandler {
	return &VerifyEmailHandler{
		userRepo:         userRepo,
		emailService:     emailService,
		sendWelcomeEmail: sendWelcomeEmail,
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

	if h.sendWelcomeEmail && h.emailService != nil {
		h.sendWelcomeEmailToUser(user)
	}

	return nil
}

func (h *VerifyEmailHandler) sendWelcomeEmailToUser(user *entities.User) error {
	emailBody := fmt.Sprintf(`
		<h2>Welcome to Apocapoc!</h2>
		<p>Your email has been successfully verified.</p>
		<p>You can now start tracking your habits and building better routines.</p>
		<p>If you have any questions or need help, please don't hesitate to contact us.</p>
	`)

	message := services.EmailMessage{
		To:      user.Email,
		Subject: "Welcome to Apocapoc!",
		Body:    emailBody,
		IsHTML:  true,
	}

	return h.emailService.Send(message)
}
