package queries

import (
	"context"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type GetHabitByIDQuery struct {
	HabitID string
	UserID  string
}

type GetHabitByIDHandler struct {
	habitRepo repositories.HabitRepository
}

func NewGetHabitByIDHandler(habitRepo repositories.HabitRepository) *GetHabitByIDHandler {
	return &GetHabitByIDHandler{
		habitRepo: habitRepo,
	}
}

func (h *GetHabitByIDHandler) Handle(ctx context.Context, query GetHabitByIDQuery) (*HabitDTO, error) {
	habit, err := h.habitRepo.FindByID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	if habit.UserID != query.UserID {
		return nil, errors.ErrUnauthorized
	}

	return &HabitDTO{
		ID:           habit.ID,
		Name:         habit.Name,
		Type:         habit.Type,
		Frequency:    habit.Frequency,
		TargetValue:  habit.TargetValue,
		CarryOver:    habit.CarryOver,
		IsNegative:   habit.IsNegative,
		SpecificDays: habit.SpecificDays,
	}, nil
}
