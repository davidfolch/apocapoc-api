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
	// Verify habit exists and user owns it
	habit, err := h.habitRepo.FindByID(ctx, cmd.HabitID)
	if err != nil {
		return err
	}

	if habit.UserID != cmd.UserID {
		return errors.ErrUnauthorized
	}

	// Find the entry for the scheduled date
	// We search within the same day
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

	// Find the entry for this date
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

	// Hard delete the entry
	return h.entryRepo.Delete(ctx, targetEntryID)
}
