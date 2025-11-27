package commands

import (
	"context"
	"fmt"
	"math"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
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

	if habit.Type == value_objects.HabitTypeCounter && cmd.Value != nil {
		if *cmd.Value != math.Floor(*cmd.Value) {
			return errors.ErrInvalidInput
		}
	}

	var finalValue *float64

	if habit.Type == value_objects.HabitTypeCounter {
		startOfDay := time.Date(cmd.ScheduledDate.Year(), cmd.ScheduledDate.Month(), cmd.ScheduledDate.Day(), 0, 0, 0, 0, cmd.ScheduledDate.Location())
		endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

		existingEntries, _ := h.entryRepo.FindByHabitIDAndDateRange(ctx, cmd.HabitID, startOfDay, endOfDay)

		if len(existingEntries) > 0 {
			existingEntry := existingEntries[0]

			var increment float64 = 1.0
			if cmd.Value != nil {
				increment = *cmd.Value
			}

			var newValue float64
			if existingEntry.Value != nil {
				newValue = *existingEntry.Value + increment
			} else {
				newValue = increment
			}

			if newValue < 0 {
				newValue = 0
			}

			finalValue = &newValue
			existingEntry.Value = finalValue
			existingEntry.CompletedAt = time.Now()

			return h.entryRepo.Update(ctx, existingEntry)
		} else {
			if cmd.Value != nil {
				value := *cmd.Value
				if value < 0 {
					value = 0
				}
				finalValue = &value
			} else {
				defaultValue := 1.0
				finalValue = &defaultValue
			}
		}
	} else {
		finalValue = cmd.Value
	}

	entry := entities.NewHabitEntry(cmd.HabitID, cmd.ScheduledDate, finalValue)

	return h.entryRepo.Create(ctx, entry)
}
