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
	"apocapoc-api/internal/shared/validation"
)

type RegisterUserCommand struct {
	Email    string
	Password string
}

type RegisterUserResult struct {
	UserID                    string
	EmailVerificationRequired bool
}

type RegisterUserHandler struct {
	userRepo         repositories.UserRepository
	passwordHasher   services.PasswordHasher
	emailService     services.EmailService
	appURL           string
	registrationMode string
	sendWelcomeEmail bool
}

func NewRegisterUserHandler(
	userRepo repositories.UserRepository,
	passwordHasher services.PasswordHasher,
	emailService services.EmailService,
	appURL string,
	registrationMode string,
	sendWelcomeEmail bool,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:         userRepo,
		passwordHasher:   passwordHasher,
		emailService:     emailService,
		appURL:           appURL,
		registrationMode: registrationMode,
		sendWelcomeEmail: sendWelcomeEmail,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResult, error) {
	if h.registrationMode == "closed" {
		return nil, errors.ErrRegistrationClosed
	}

	if err := validation.ValidateRegistration(cmd.Email, cmd.Password); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrInvalidInput, err)
	}

	existing, _ := h.userRepo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return nil, errors.ErrAlreadyExists
	}

	hashedPassword, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	user := entities.NewUser(cmd.Email, hashedPassword)

	emailVerificationRequired := false
	if h.emailService != nil {
		token, err := h.generateVerificationToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate verification token: %w", err)
		}

		expiry := time.Now().Add(24 * time.Hour)
		user.EmailVerificationToken = &token
		user.EmailVerificationExpiry = &expiry
		emailVerificationRequired = true

		if err := h.sendVerificationEmail(user); err != nil {
			return nil, fmt.Errorf("failed to send verification email: %w", err)
		}
	} else {
		user.EmailVerified = true
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterUserResult{
		UserID:                    user.ID,
		EmailVerificationRequired: emailVerificationRequired,
	}, nil
}

func (h *RegisterUserHandler) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *RegisterUserHandler) sendVerificationEmail(user *entities.User) error {
	if user.EmailVerificationToken == nil {
		return nil
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", h.appURL, *user.EmailVerificationToken)

	emailBody := fmt.Sprintf(`
		<h2>Welcome! Please verify your email</h2>
		<p>Thank you for registering. Please click the link below to verify your email address:</p>
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

	return h.emailService.Send(message)
}
