package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
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
	From    *time.Time
	To      *time.Time
	Page    int
	Limit   int
}

type GetHabitEntriesResult struct {
	Entries []HabitEntryDTO
	Total   int
	Page    int
	Limit   int
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

func (h *GetHabitEntriesHandler) Handle(ctx context.Context, query GetHabitEntriesQuery) (*GetHabitEntriesResult, error) {
	habit, err := h.habitRepo.FindByID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	if habit.UserID != query.UserID {
		return nil, errors.ErrUnauthorized
	}

	var dateRangeDays int
	if query.From != nil && query.To != nil {
		dateRangeDays = int(query.To.Sub(*query.From).Hours() / 24)
	}

	requiresPagination := false
	if query.From == nil || query.To == nil {
		requiresPagination = true
	} else if dateRangeDays > 365 {
		requiresPagination = true
	}

	if requiresPagination && query.Limit == 0 {
		return nil, errors.ErrInvalidInput
	}

	var entries []*entities.HabitEntry

	if query.From != nil && query.To != nil {
		entries, err = h.entryRepo.FindByHabitIDAndDateRange(ctx, query.HabitID, *query.From, *query.To)
	} else if query.From != nil {
		entries, err = h.entryRepo.FindByHabitIDAndDateRange(ctx, query.HabitID, *query.From, time.Now())
	} else if query.To != nil {
		entries, err = h.entryRepo.FindByHabitIDAndDateRange(ctx, query.HabitID, time.Time{}, *query.To)
	} else {
		entries, err = h.entryRepo.FindByHabitID(ctx, query.HabitID)
	}

	if err != nil {
		return nil, err
	}

	total := len(entries)

	if query.Limit > 0 {
		offset := (query.Page - 1) * query.Limit
		if offset < 0 {
			offset = 0
		}
		end := offset + query.Limit
		if offset < len(entries) {
			if end > len(entries) {
				end = len(entries)
			}
			entries = entries[offset:end]
		} else {
			entries = []*entities.HabitEntry{}
		}
	}

	var result []HabitEntryDTO
	for _, entry := range entries {
		result = append(result, HabitEntryDTO{
			ID:            entry.ID,
			HabitID:       entry.HabitID,
			ScheduledDate: entry.ScheduledDate,
			CompletedAt:   entry.CompletedAt,
			Value:         entry.Value,
		})
	}

	return &GetHabitEntriesResult{
		Entries: result,
		Total:   total,
		Page:    query.Page,
		Limit:   query.Limit,
	}, nil
}
