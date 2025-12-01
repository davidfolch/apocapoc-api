package repositories

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/pagination"
)

type HabitRepository interface {
	Create(ctx context.Context, habit *entities.Habit) error
	FindByID(ctx context.Context, id string) (*entities.Habit, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error)
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
	Update(ctx context.Context, habit *entities.Habit) error
	Delete(ctx context.Context, id string) error
}
