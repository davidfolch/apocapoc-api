package repositories

import (
	"context"
	"time"

	"habit-tracker-api/internal/domain/entities"
)

type HabitEntryRepository interface {
	Create(ctx context.Context, entry *entities.HabitEntry) error
	FindByID(ctx context.Context, id string) (*entities.HabitEntry, error)
	FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error)
	FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error)
	FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error)
	Update(ctx context.Context, entry *entities.HabitEntry) error
	Delete(ctx context.Context, id string) error
}
