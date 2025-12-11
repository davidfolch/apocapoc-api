package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type HabitChangesDTO struct {
	Created []*entities.Habit
	Updated []*entities.Habit
	Deleted []string
}

type EntryChangesDTO struct {
	Created []*entities.HabitEntry
	Updated []*entities.HabitEntry
	Deleted []string
}

type SyncChangesDTO struct {
	Habits  HabitChangesDTO
	Entries EntryChangesDTO
}

type GetSyncChangesQuery struct {
	UserID string
	Since  time.Time
}

type GetSyncChangesHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewGetSyncChangesHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *GetSyncChangesHandler {
	return &GetSyncChangesHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *GetSyncChangesHandler) Handle(ctx context.Context, query GetSyncChangesQuery) (*SyncChangesDTO, error) {
	if query.UserID == "" {
		return nil, errors.ErrInvalidInput
	}

	habitChanges, err := h.habitRepo.GetChangesSince(ctx, query.UserID, query.Since)
	if err != nil {
		return nil, err
	}

	entryChanges, err := h.entryRepo.GetChangesSince(ctx, query.UserID, query.Since)
	if err != nil {
		return nil, err
	}

	return &SyncChangesDTO{
		Habits: HabitChangesDTO{
			Created: habitChanges.Created,
			Updated: habitChanges.Updated,
			Deleted: habitChanges.Deleted,
		},
		Entries: EntryChangesDTO{
			Created: entryChanges.Created,
			Updated: entryChanges.Updated,
			Deleted: entryChanges.Deleted,
		},
	}, nil
}
