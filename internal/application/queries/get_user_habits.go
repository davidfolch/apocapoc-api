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

type FilterParams struct {
	Type            *value_objects.HabitType
	Frequency       *value_objects.Frequency
	IncludeArchived bool
	Search          string
}

type GetUserHabitsQuery struct {
	UserID           string
	PaginationParams *pagination.Params
	FilterParams     *FilterParams
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

	if query.FilterParams != nil {
		filter := repositories.HabitFilter{
			Type:            query.FilterParams.Type,
			Frequency:       query.FilterParams.Frequency,
			IncludeArchived: query.FilterParams.IncludeArchived,
			Search:          query.FilterParams.Search,
		}

		habits, err = h.habitRepo.FindByUserIDFiltered(ctx, query.UserID, filter, query.PaginationParams)
		if err != nil {
			return nil, err
		}

		if query.PaginationParams != nil {
			totalItems, err := h.habitRepo.CountByUserIDFiltered(ctx, query.UserID, filter)
			if err != nil {
				return nil, err
			}

			response := pagination.NewResponse(*query.PaginationParams, totalItems)
			paginationResponse = &response
		}
	} else if query.PaginationParams != nil {
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
