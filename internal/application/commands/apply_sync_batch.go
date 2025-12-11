package commands

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type HabitBatchChanges struct {
	Created []*entities.Habit
	Updated []*entities.Habit
	Deleted []string
}

type EntryBatchChanges struct {
	Created []*entities.HabitEntry
	Updated []*entities.HabitEntry
	Deleted []string
}

type ApplySyncBatchCommand struct {
	UserID  string
	Habits  HabitBatchChanges
	Entries EntryBatchChanges
}

type ApplySyncBatchHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewApplySyncBatchHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *ApplySyncBatchHandler {
	return &ApplySyncBatchHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *ApplySyncBatchHandler) Handle(ctx context.Context, cmd ApplySyncBatchCommand) error {
	if cmd.UserID == "" {
		return errors.ErrInvalidInput
	}

	for _, habit := range cmd.Habits.Created {
		if habit.UserID != cmd.UserID {
			return errors.ErrUnauthorized
		}
		if err := h.habitRepo.Create(ctx, habit); err != nil {
			return err
		}
	}

	for _, habit := range cmd.Habits.Updated {
		if habit.UserID != cmd.UserID {
			return errors.ErrUnauthorized
		}

		existing, err := h.habitRepo.FindByID(ctx, habit.ID)
		if err != nil {
			if err == errors.ErrNotFound {
				if err := h.habitRepo.Create(ctx, habit); err != nil {
					return err
				}
				continue
			}
			return err
		}

		if existing.UserID != cmd.UserID {
			return errors.ErrUnauthorized
		}

		if shouldApplyUpdate(existing.UpdatedAt, habit.UpdatedAt) {
			if err := h.habitRepo.Update(ctx, habit); err != nil {
				return err
			}
		}
	}

	for _, id := range cmd.Habits.Deleted {
		existing, err := h.habitRepo.FindByID(ctx, id)
		if err != nil {
			if err == errors.ErrNotFound {
				continue
			}
			return err
		}

		if existing.UserID != cmd.UserID {
			return errors.ErrUnauthorized
		}

		if err := h.habitRepo.SoftDelete(ctx, id); err != nil && err != errors.ErrNotFound {
			return err
		}
	}

	for _, entry := range cmd.Entries.Created {
		if err := h.entryRepo.Create(ctx, entry); err != nil {
			return err
		}
	}

	for _, entry := range cmd.Entries.Updated {
		existing, err := h.entryRepo.FindByID(ctx, entry.ID)
		if err != nil {
			if err == errors.ErrNotFound {
				if err := h.entryRepo.Create(ctx, entry); err != nil {
					return err
				}
				continue
			}
			return err
		}

		if shouldApplyUpdate(existing.UpdatedAt, entry.UpdatedAt) {
			if err := h.entryRepo.Update(ctx, entry); err != nil {
				return err
			}
		}
	}

	for _, id := range cmd.Entries.Deleted {
		if err := h.entryRepo.SoftDelete(ctx, id); err != nil && err != errors.ErrNotFound {
			return err
		}
	}

	return nil
}

func shouldApplyUpdate(serverTime, clientTime time.Time) bool {
	return clientTime.After(serverTime)
}
