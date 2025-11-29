package queries

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type LoginUserQuery struct {
	Email    string
	Password string
}

type LoginUserResult struct {
	UserID string
	Email  string
}

type LoginUserHandler struct {
	userRepo       repositories.UserRepository
	passwordHasher services.PasswordHasher
}

func NewLoginUserHandler(userRepo repositories.UserRepository, passwordHasher services.PasswordHasher) *LoginUserHandler {
	return &LoginUserHandler{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, query LoginUserQuery) (*LoginUserResult, error) {
	if query.Email == "" || query.Password == "" {
		return nil, errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByEmail(ctx, query.Email)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if err := h.passwordHasher.Compare(user.PasswordHash, query.Password); err != nil {
		return nil, errors.ErrNotFound
	}

	if !user.EmailVerified {
		return nil, errors.ErrEmailNotVerified
	}

	return &LoginUserResult{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}
