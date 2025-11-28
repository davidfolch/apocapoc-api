package commands

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type DeleteUserCommand struct {
	UserID string
}

type DeleteUserHandler struct {
	userRepo repositories.UserRepository
}

func NewDeleteUserHandler(userRepo repositories.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{
		userRepo: userRepo,
	}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) error {
	if cmd.UserID == "" {
		return errors.ErrInvalidInput
	}

	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return errors.ErrNotFound
	}

	if err := h.userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	return nil
}
