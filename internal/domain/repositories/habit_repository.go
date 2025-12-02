package repositories

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/pagination"
)

type HabitFilter struct {
	Type            *value_objects.HabitType
	Frequency       *value_objects.Frequency
	IncludeArchived bool
	Search          string
}

type HabitRepository interface {
	Create(ctx context.Context, habit *entities.Habit) error
	FindByID(ctx context.Context, id string) (*entities.Habit, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error)
	FindByUserIDFiltered(ctx context.Context, userID string, filter HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error)
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
	CountByUserIDFiltered(ctx context.Context, userID string, filter HabitFilter) (int, error)
	Update(ctx context.Context, habit *entities.Habit) error
	Delete(ctx context.Context, id string) error
}
