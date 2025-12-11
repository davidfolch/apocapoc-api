package queries

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/pagination"
)

type mockHabitRepoForSync struct {
	changes *repositories.HabitChanges
	err     error
}

func (m *mockHabitRepoForSync) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	return m.changes, m.err
}

func (m *mockHabitRepoForSync) Create(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepoForSync) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForSync) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForSync) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForSync) Update(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepoForSync) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockHabitRepoForSync) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForSync) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForSync) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForSync) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForSync) SoftDelete(ctx context.Context, id string) error {
	return nil
}

type mockEntryRepoForSync struct {
	changes *repositories.HabitEntryChanges
	err     error
}

func (m *mockEntryRepoForSync) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitEntryChanges, error) {
	return m.changes, m.err
}

func (m *mockEntryRepoForSync) Create(ctx context.Context, entry *entities.HabitEntry) error {
	return nil
}

func (m *mockEntryRepoForSync) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForSync) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForSync) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForSync) FindByUserID(ctx context.Context, userID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForSync) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForSync) Update(ctx context.Context, entry *entities.HabitEntry) error {
	return nil
}

func (m *mockEntryRepoForSync) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockEntryRepoForSync) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func TestGetSyncChangesHandler_Success(t *testing.T) {
	now := time.Now()
	since := now.Add(-1 * time.Hour)

	createdHabit := entities.NewHabit("user-123", "New Habit", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	createdHabit.ID = "habit-1"
	createdHabit.CreatedAt = now
	createdHabit.UpdatedAt = now

	updatedHabit := entities.NewHabit("user-123", "Updated Habit", value_objects.HabitTypeCounter, value_objects.FrequencyWeekly, false, false)
	updatedHabit.ID = "habit-2"
	updatedHabit.CreatedAt = since.Add(-1 * time.Hour)
	updatedHabit.UpdatedAt = now

	habitRepo := &mockHabitRepoForSync{
		changes: &repositories.HabitChanges{
			Created: []*entities.Habit{createdHabit},
			Updated: []*entities.Habit{updatedHabit},
			Deleted: []string{"habit-3"},
		},
	}

	createdEntry := entities.NewHabitEntry("habit-1", now, nil)
	createdEntry.ID = "entry-1"

	updatedEntry := entities.NewHabitEntry("habit-2", now, nil)
	updatedEntry.ID = "entry-2"
	updatedEntry.UpdatedAt = now

	entryRepo := &mockEntryRepoForSync{
		changes: &repositories.HabitEntryChanges{
			Created: []*entities.HabitEntry{createdEntry},
			Updated: []*entities.HabitEntry{updatedEntry},
			Deleted: []string{"entry-3"},
		},
	}

	handler := NewGetSyncChangesHandler(habitRepo, entryRepo)

	query := GetSyncChangesQuery{
		UserID: "user-123",
		Since:  since,
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Habits.Created) != 1 {
		t.Errorf("Expected 1 created habit, got %d", len(result.Habits.Created))
	}

	if len(result.Habits.Updated) != 1 {
		t.Errorf("Expected 1 updated habit, got %d", len(result.Habits.Updated))
	}

	if len(result.Habits.Deleted) != 1 {
		t.Errorf("Expected 1 deleted habit, got %d", len(result.Habits.Deleted))
	}

	if len(result.Entries.Created) != 1 {
		t.Errorf("Expected 1 created entry, got %d", len(result.Entries.Created))
	}

	if len(result.Entries.Updated) != 1 {
		t.Errorf("Expected 1 updated entry, got %d", len(result.Entries.Updated))
	}

	if len(result.Entries.Deleted) != 1 {
		t.Errorf("Expected 1 deleted entry, got %d", len(result.Entries.Deleted))
	}
}

func TestGetSyncChangesHandler_EmptyChanges(t *testing.T) {
	habitRepo := &mockHabitRepoForSync{
		changes: &repositories.HabitChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{},
			Deleted: []string{},
		},
	}

	entryRepo := &mockEntryRepoForSync{
		changes: &repositories.HabitEntryChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	handler := NewGetSyncChangesHandler(habitRepo, entryRepo)

	query := GetSyncChangesQuery{
		UserID: "user-123",
		Since:  time.Now().Add(-1 * time.Hour),
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Habits.Created) != 0 {
		t.Errorf("Expected 0 created habits, got %d", len(result.Habits.Created))
	}

	if len(result.Entries.Created) != 0 {
		t.Errorf("Expected 0 created entries, got %d", len(result.Entries.Created))
	}
}

func TestGetSyncChangesHandler_InvalidUserID(t *testing.T) {
	habitRepo := &mockHabitRepoForSync{
		changes: &repositories.HabitChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{},
			Deleted: []string{},
		},
	}

	entryRepo := &mockEntryRepoForSync{
		changes: &repositories.HabitEntryChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	handler := NewGetSyncChangesHandler(habitRepo, entryRepo)

	query := GetSyncChangesQuery{
		UserID: "",
		Since:  time.Now(),
	}

	_, err := handler.Handle(context.Background(), query)

	if err == nil {
		t.Error("Expected error for empty UserID, got nil")
	}
}
