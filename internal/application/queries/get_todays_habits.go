package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/utils"
)

type TodaysHabitDTO struct {
	ID            string
	Name          string
	Type          string
	TargetValue   *float64
	ScheduledDate time.Time
	IsCarriedOver bool
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
			query.Date.AddDate(0, 0, -30),
			query.Date,
		)

		isCompleted := false
		for _, entry := range entries {
			if entry.ScheduledDate.Equal(query.Date) && entry.DeletedAt == nil {
				isCompleted = true
				break
			}
		}

		if !isCompleted {
			result = append(result, TodaysHabitDTO{
				ID:            habit.ID,
				Name:          habit.Name,
				Type:          string(habit.Type),
				TargetValue:   habit.TargetValue,
				ScheduledDate: query.Date,
				IsCarriedOver: !shouldAppear && habit.CarryOver,
			})
		}
	}

	return result, nil
}
