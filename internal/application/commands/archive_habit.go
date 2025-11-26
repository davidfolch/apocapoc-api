package commands

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type ArchiveHabitCommand struct {
	HabitID string
	UserID  string
}

type ArchiveHabitHandler struct {
	habitRepo repositories.HabitRepository
}

func NewArchiveHabitHandler(habitRepo repositories.HabitRepository) *ArchiveHabitHandler {
	return &ArchiveHabitHandler{
		habitRepo: habitRepo,
	}
}

func (h *ArchiveHabitHandler) Handle(ctx context.Context, cmd ArchiveHabitCommand) error {
	habit, err := h.habitRepo.FindByID(ctx, cmd.HabitID)
	if err != nil {
		return err
	}

	if habit.UserID != cmd.UserID {
		return errors.ErrUnauthorized
	}

	habit.Archive()

	return h.habitRepo.Update(ctx, habit)
}
