package commands

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"

	"golang.org/x/crypto/bcrypt"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	Timezone string
}

type RegisterUserHandler struct {
	userRepo repositories.UserRepository
}

func NewRegisterUserHandler(userRepo repositories.UserRepository) *RegisterUserHandler {
	return &RegisterUserHandler{userRepo: userRepo}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (string, error) {
	if cmd.Email == "" || cmd.Password == "" {
		return "", errors.ErrInvalidInput
	}

	if len(cmd.Password) < 8 {
		return "", errors.ErrInvalidInput
	}

	existing, _ := h.userRepo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return "", errors.ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	timezone := cmd.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	user := entities.NewUser(cmd.Email, string(hashedPassword), timezone)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return user.ID, nil
}
