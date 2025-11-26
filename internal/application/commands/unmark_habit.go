package commands

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type UnmarkHabitCommand struct {
	HabitID       string
	UserID        string
	ScheduledDate time.Time
}

type UnmarkHabitHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewUnmarkHabitHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *UnmarkHabitHandler {
	return &UnmarkHabitHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *UnmarkHabitHandler) Handle(ctx context.Context, cmd UnmarkHabitCommand) error {
	habit, err := h.habitRepo.FindByID(ctx, cmd.HabitID)
	if err != nil {
		return err
	}

	if habit.UserID != cmd.UserID {
		return errors.ErrUnauthorized
	}

	startOfDay := time.Date(
		cmd.ScheduledDate.Year(),
		cmd.ScheduledDate.Month(),
		cmd.ScheduledDate.Day(),
		0, 0, 0, 0,
		cmd.ScheduledDate.Location(),
	)
	endOfDay := startOfDay.Add(24 * time.Hour)

	entries, err := h.entryRepo.FindByHabitIDAndDateRange(ctx, cmd.HabitID, startOfDay, endOfDay)
	if err != nil {
		return err
	}

	var targetEntryID string
	for _, entry := range entries {
		if entry.ScheduledDate.Equal(cmd.ScheduledDate) {
			targetEntryID = entry.ID
			break
		}
	}

	if targetEntryID == "" {
		return errors.ErrNotFound
	}

	return h.entryRepo.Delete(ctx, targetEntryID)
}
