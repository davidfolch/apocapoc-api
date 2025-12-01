package queries

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/pagination"
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
	UserID           string
	PaginationParams *pagination.Params
}

type GetUserHabitsResult struct {
	Habits     []HabitDTO
	Pagination *pagination.Response
}

type GetUserHabitsHandler struct {
	habitRepo repositories.HabitRepository
}

func NewGetUserHabitsHandler(habitRepo repositories.HabitRepository) *GetUserHabitsHandler {
	return &GetUserHabitsHandler{
		habitRepo: habitRepo,
	}
}

func (h *GetUserHabitsHandler) Handle(ctx context.Context, query GetUserHabitsQuery) (*GetUserHabitsResult, error) {
	var habits []*entities.Habit
	var paginationResponse *pagination.Response
	var err error

	if query.PaginationParams != nil {
		habits, err = h.habitRepo.FindActiveByUserIDWithPagination(ctx, query.UserID, *query.PaginationParams)
		if err != nil {
			return nil, err
		}

		totalItems, err := h.habitRepo.CountActiveByUserID(ctx, query.UserID)
		if err != nil {
			return nil, err
		}

		response := pagination.NewResponse(*query.PaginationParams, totalItems)
		paginationResponse = &response
	} else {
		habits, err = h.habitRepo.FindActiveByUserID(ctx, query.UserID)
		if err != nil {
			return nil, err
		}
	}

	var habitDTOs []HabitDTO
	for _, habit := range habits {
		habitDTOs = append(habitDTOs, HabitDTO{
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

	return &GetUserHabitsResult{
		Habits:     habitDTOs,
		Pagination: paginationResponse,
	}, nil
}
