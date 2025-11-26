package queries

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"

	"golang.org/x/crypto/bcrypt"
)

type LoginUserQuery struct {
	Email    string
	Password string
}

type LoginUserResult struct {
	UserID   string
	Email    string
	Timezone string
}

type LoginUserHandler struct {
	userRepo repositories.UserRepository
}

func NewLoginUserHandler(userRepo repositories.UserRepository) *LoginUserHandler {
	return &LoginUserHandler{userRepo: userRepo}
}

func (h *LoginUserHandler) Handle(ctx context.Context, query LoginUserQuery) (*LoginUserResult, error) {
	if query.Email == "" || query.Password == "" {
		return nil, errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByEmail(ctx, query.Email)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(query.Password)); err != nil {
		return nil, errors.ErrNotFound
	}

	return &LoginUserResult{
		UserID:   user.ID,
		Email:    user.Email,
		Timezone: user.Timezone,
	}, nil
}
