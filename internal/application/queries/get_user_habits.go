package queries

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
)

type HabitDTO struct {
	ID           string
	Name         string
	Type         value_objects.HabitType
	Frequency    value_objects.Frequency
	TargetValue  *float64
	CarryOver    bool
	IsNegative   bool
	SpecificDays []int
}

type GetUserHabitsQuery struct {
	UserID string
}

type GetUserHabitsHandler struct {
	habitRepo repositories.HabitRepository
}

func NewGetUserHabitsHandler(habitRepo repositories.HabitRepository) *GetUserHabitsHandler {
	return &GetUserHabitsHandler{
		habitRepo: habitRepo,
	}
}

func (h *GetUserHabitsHandler) Handle(ctx context.Context, query GetUserHabitsQuery) ([]HabitDTO, error) {
	habits, err := h.habitRepo.FindActiveByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	var result []HabitDTO
	for _, habit := range habits {
		result = append(result, HabitDTO{
			ID:           habit.ID,
			Name:         habit.Name,
			Type:         habit.Type,
			Frequency:    habit.Frequency,
			TargetValue:  habit.TargetValue,
			CarryOver:    habit.CarryOver,
			IsNegative:   habit.IsNegative,
			SpecificDays: habit.SpecificDays,
		})
	}

	return result, nil
}
