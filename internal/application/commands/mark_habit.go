package commands

import (
	"context"
	"fmt"
	"time"

	"habit-tracker-api/internal/domain/entities"
	"habit-tracker-api/internal/domain/repositories"
)

type MarkHabitCommand struct {
	HabitID       string
	ScheduledDate time.Time
	Value         *float64
}

type MarkHabitHandler struct {
	entryRepo repositories.HabitEntryRepository
	habitRepo repositories.HabitRepository
}

func NewMarkHabitHandler(
	entryRepo repositories.HabitEntryRepository,
	habitRepo repositories.HabitRepository,
) *MarkHabitHandler {
	return &MarkHabitHandler{
		entryRepo: entryRepo,
		habitRepo: habitRepo,
	}
}

func (h *MarkHabitHandler) Handle(ctx context.Context, cmd MarkHabitCommand) error {
	habit, err := h.habitRepo.FindByID(ctx, cmd.HabitID)
	if err != nil {
		return err
	}

	if habit == nil {
		return fmt.Errorf("habit not found")
	}

	if !habit.IsActive() {
		return fmt.Errorf("habit is archived")
	}

	entry := entities.NewHabitEntry(cmd.HabitID, cmd.ScheduledDate, cmd.Value)

	return h.entryRepo.Create(ctx, entry)
}
