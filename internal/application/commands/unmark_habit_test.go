package commands

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type mockEntryRepoForUnmark struct {
	mockEntryRepo
	entries       []*entities.HabitEntry
	updatedEntry  *entities.HabitEntry
	errorOnUpdate error
}

func (m *mockEntryRepoForUnmark) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
	return m.entries, nil
}

func (m *mockEntryRepoForUnmark) Update(ctx context.Context, entry *entities.HabitEntry) error {
	if m.errorOnUpdate != nil {
		return m.errorOnUpdate
	}
	m.updatedEntry = entry
	return nil
}

func TestUnmarkHabitHandler_UnmarksSuccessfully(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	scheduledDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	entry := entities.NewHabitEntry("habit-1", scheduledDate, nil)
	entry.ID = "entry-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoForUnmark{
		entries: []*entities.HabitEntry{entry},
	}

	handler := NewUnmarkHabitHandler(habitRepo, entryRepo)

	cmd := UnmarkHabitCommand{
		HabitID:       "habit-1",
		UserID:        "user-123",
		ScheduledDate: scheduledDate,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entryRepo.updatedEntry == nil {
		t.Fatal("Expected entry to be updated")
	}

	if entryRepo.updatedEntry.DeletedAt == nil {
		t.Error("Expected entry to be soft deleted")
	}
}

func TestUnmarkHabitHandler_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoForUpdate{
		errorOnFind: errors.ErrNotFound,
	}

	entryRepo := &mockEntryRepoForUnmark{}

	handler := NewUnmarkHabitHandler(habitRepo, entryRepo)

	cmd := UnmarkHabitCommand{
		HabitID:       "non-existent",
		UserID:        "user-123",
		ScheduledDate: time.Now(),
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestUnmarkHabitHandler_ReturnsErrorWhenUserDoesNotOwnHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoForUnmark{}

	handler := NewUnmarkHabitHandler(habitRepo, entryRepo)

	cmd := UnmarkHabitCommand{
		HabitID:       "habit-1",
		UserID:        "user-456", // Different user
		ScheduledDate: time.Now(),
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestUnmarkHabitHandler_ReturnsErrorWhenEntryNotFound(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	// No entries
	entryRepo := &mockEntryRepoForUnmark{
		entries: []*entities.HabitEntry{},
	}

	handler := NewUnmarkHabitHandler(habitRepo, entryRepo)

	cmd := UnmarkHabitCommand{
		HabitID:       "habit-1",
		UserID:        "user-123",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound for missing entry, got %v", err)
	}
}

func TestUnmarkHabitHandler_IgnoresAlreadyDeletedEntry(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	scheduledDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	entry := entities.NewHabitEntry("habit-1", scheduledDate, nil)
	entry.ID = "entry-1"
	now := time.Now()
	entry.DeletedAt = &now // Already deleted

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoForUnmark{
		entries: []*entities.HabitEntry{entry},
	}

	handler := NewUnmarkHabitHandler(habitRepo, entryRepo)

	cmd := UnmarkHabitCommand{
		HabitID:       "habit-1",
		UserID:        "user-123",
		ScheduledDate: scheduledDate,
	}

	err := handler.Handle(context.Background(), cmd)

	// Should return not found since the active entry doesn't exist
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound for already deleted entry, got %v", err)
	}
}
