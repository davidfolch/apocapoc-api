package queries

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
)

type HabitDTO struct {
	ID           string
	Name         string
	Type         string
	Frequency    string
	TargetValue  *float64
	CarryOver    bool
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
			Type:         string(habit.Type),
			Frequency:    string(habit.Frequency),
			TargetValue:  habit.TargetValue,
			CarryOver:    habit.CarryOver,
			SpecificDays: habit.SpecificDays,
		})
	}

	return result, nil
}
