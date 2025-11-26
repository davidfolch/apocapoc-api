package commands

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
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

	// Find the active entry for this date
	var targetEntry *entities.HabitEntry
	for _, entry := range entries {
		if entry.ScheduledDate.Equal(cmd.ScheduledDate) && entry.DeletedAt == nil {
			targetEntry = entry
			break
		}
	}

	if targetEntry == nil {
		return errors.ErrNotFound
	}

	// Soft delete the entry
	now := time.Now()
	targetEntry.DeletedAt = &now

	return h.entryRepo.Update(ctx, targetEntry)
}
