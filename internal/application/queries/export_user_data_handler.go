package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
)

type ExportHabitDTO struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	Type          value_objects.HabitType  `json:"type"`
	Frequency     value_objects.Frequency  `json:"frequency"`
	SpecificDays  []int                    `json:"specific_days,omitempty"`
	SpecificDates []int                    `json:"specific_dates,omitempty"`
	CarryOver     bool                     `json:"carry_over"`
	IsNegative    bool                     `json:"is_negative"`
	TargetValue   *float64                 `json:"target_value,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
	ArchivedAt    *time.Time               `json:"archived_at,omitempty"`
}

type ExportEntryDTO struct {
	ID            string     `json:"id"`
	HabitID       string     `json:"habit_id"`
	ScheduledDate time.Time  `json:"scheduled_date"`
	CompletedAt   time.Time  `json:"completed_at"`
	Value         *float64   `json:"value,omitempty"`
}

type ExportUserDataResult struct {
	ExportedAt time.Time        `json:"exported_at"`
	Habits     []ExportHabitDTO `json:"habits"`
	Entries    []ExportEntryDTO `json:"entries"`
}

type ExportUserDataQuery struct {
	UserID string
}

type ExportUserDataHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewExportUserDataHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *ExportUserDataHandler {
	return &ExportUserDataHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *ExportUserDataHandler) Handle(ctx context.Context, query ExportUserDataQuery) (*ExportUserDataResult, error) {
	habits, err := h.habitRepo.FindByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	entries, err := h.entryRepo.FindByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	habitDTOs := make([]ExportHabitDTO, 0, len(habits))
	for _, habit := range habits {
		habitDTOs = append(habitDTOs, ExportHabitDTO{
			ID:            habit.ID,
			Name:          habit.Name,
			Description:   habit.Description,
			Type:          habit.Type,
			Frequency:     habit.Frequency,
			SpecificDays:  habit.SpecificDays,
			SpecificDates: habit.SpecificDates,
			CarryOver:     habit.CarryOver,
			IsNegative:    habit.IsNegative,
			TargetValue:   habit.TargetValue,
			CreatedAt:     habit.CreatedAt,
			ArchivedAt:    habit.ArchivedAt,
		})
	}

	entryDTOs := make([]ExportEntryDTO, 0, len(entries))
	for _, entry := range entries {
		entryDTOs = append(entryDTOs, ExportEntryDTO{
			ID:            entry.ID,
			HabitID:       entry.HabitID,
			ScheduledDate: entry.ScheduledDate,
			CompletedAt:   entry.CompletedAt,
			Value:         entry.Value,
		})
	}

	return &ExportUserDataResult{
		ExportedAt: time.Now(),
		Habits:     habitDTOs,
		Entries:    entryDTOs,
	}, nil
}
