package commands

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
	"apocapoc-api/internal/shared/validation"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	Timezone string
}

type RegisterUserHandler struct {
	userRepo       repositories.UserRepository
	passwordHasher services.PasswordHasher
}

func NewRegisterUserHandler(userRepo repositories.UserRepository, passwordHasher services.PasswordHasher) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (string, error) {
	if err := validation.ValidateRegistration(cmd.Email, cmd.Password, cmd.Timezone); err != nil {
		return "", errors.ErrInvalidInput
	}

	existing, _ := h.userRepo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return "", errors.ErrAlreadyExists
	}

	hashedPassword, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return "", err
	}

	user := entities.NewUser(cmd.Email, hashedPassword, cmd.Timezone)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return user.ID, nil
}
