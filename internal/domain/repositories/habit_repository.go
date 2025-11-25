package repositories

import (
	"context"

	"habit-tracker-api/internal/domain/entities"
)

type HabitRepository interface {
	Create(ctx context.Context, habit *entities.Habit) error
	FindByID(ctx context.Context, id string) (*entities.Habit, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error)
	Update(ctx context.Context, habit *entities.Habit) error
	Delete(ctx context.Context, id string) error
}
