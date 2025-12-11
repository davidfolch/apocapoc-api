package repositories

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
)

type HabitEntryChanges struct {
	Created []*entities.HabitEntry
	Updated []*entities.HabitEntry
	Deleted []string
}

type HabitEntryRepository interface {
	Create(ctx context.Context, entry *entities.HabitEntry) error
	FindByID(ctx context.Context, id string) (*entities.HabitEntry, error)
	FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error)
	FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.HabitEntry, error)
	FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error)
	Update(ctx context.Context, entry *entities.HabitEntry) error
	Delete(ctx context.Context, id string) error

	// Sync methods
	GetChangesSince(ctx context.Context, userID string, since time.Time) (*HabitEntryChanges, error)
	SoftDelete(ctx context.Context, id string) error
}
