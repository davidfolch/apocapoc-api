package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type HabitEntryDTO struct {
	ID            string
	HabitID       string
	ScheduledDate time.Time
	CompletedAt   time.Time
	Value         *float64
}

type GetHabitEntriesQuery struct {
	HabitID string
	UserID  string
}

type GetHabitEntriesHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewGetHabitEntriesHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *GetHabitEntriesHandler {
	return &GetHabitEntriesHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *GetHabitEntriesHandler) Handle(ctx context.Context, query GetHabitEntriesQuery) ([]HabitEntryDTO, error) {
	// Verify habit exists and user owns it
	habit, err := h.habitRepo.FindByID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	if habit.UserID != query.UserID {
		return nil, errors.ErrUnauthorized
	}

	// Get all entries for the habit
	entries, err := h.entryRepo.FindByHabitID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	var result []HabitEntryDTO
	for _, entry := range entries {
		// Filter out deleted entries
		if entry.DeletedAt != nil {
			continue
		}

		result = append(result, HabitEntryDTO{
			ID:            entry.ID,
			HabitID:       entry.HabitID,
			ScheduledDate: entry.ScheduledDate,
			CompletedAt:   entry.CompletedAt,
			Value:         entry.Value,
		})
	}

	return result, nil
}
