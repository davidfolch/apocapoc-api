package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/utils"
)

type TodaysHabitEntryDTO struct {
	ID          string
	Value       *float64
	CompletedAt time.Time
}

type TodaysHabitDTO struct {
	ID            string
	Name          string
	Type          value_objects.HabitType
	TargetValue   *float64
	IsNegative    bool
	ScheduledDate time.Time
	IsCarriedOver bool
	Entry         *TodaysHabitEntryDTO
}

type GetTodaysHabitsQuery struct {
	UserID   string
	Timezone string
	Date     time.Time
}

type GetTodaysHabitsHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewGetTodaysHabitsHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *GetTodaysHabitsHandler {
	return &GetTodaysHabitsHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *GetTodaysHabitsHandler) Handle(
	ctx context.Context,
	query GetTodaysHabitsQuery,
) ([]TodaysHabitDTO, error) {
	habits, err := h.habitRepo.FindActiveByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	var result []TodaysHabitDTO

	for _, habit := range habits {
		shouldAppear := utils.ShouldAppearToday(
			string(habit.Frequency),
			habit.SpecificDays,
			habit.SpecificDates,
			query.Date,
		)

		if !shouldAppear && !habit.CarryOver {
			continue
		}

		entries, _ := h.entryRepo.FindByHabitIDAndDateRange(
			ctx,
			habit.ID,
			query.Date,
			query.Date,
		)

		var entryDTO *TodaysHabitEntryDTO
		if len(entries) > 0 && entries[0].ScheduledDate.Format("2006-01-02") == query.Date.Format("2006-01-02") {
			entryDTO = &TodaysHabitEntryDTO{
				ID:          entries[0].ID,
				Value:       entries[0].Value,
				CompletedAt: entries[0].CompletedAt,
			}
		}

		result = append(result, TodaysHabitDTO{
			ID:            habit.ID,
			Name:          habit.Name,
			Type:          habit.Type,
			TargetValue:   habit.TargetValue,
			IsNegative:    habit.IsNegative,
			ScheduledDate: query.Date,
			IsCarriedOver: !shouldAppear && habit.CarryOver,
			Entry:         entryDTO,
		})
	}

	return result, nil
}
