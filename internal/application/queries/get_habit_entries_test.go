package queries

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type mockEntryRepoWithFindByHabitID struct {
	mockEntryRepo
	entries []*entities.HabitEntry
}

func (m *mockEntryRepoWithFindByHabitID) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	return m.entries, nil
}

func TestGetHabitEntriesHandler_ReturnsEntriesSuccessfully(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	date1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	entry1 := entities.NewHabitEntry("habit-1", date1, nil)
	entry1.ID = "entry-1"

	entry2 := entities.NewHabitEntry("habit-1", date2, nil)
	entry2.ID = "entry-2"

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: []*entities.HabitEntry{entry1, entry2},
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(results))
	}

	if results[0].ID != "entry-1" {
		t.Errorf("Expected first entry ID entry-1, got %s", results[0].ID)
	}

	if results[1].ID != "entry-2" {
		t.Errorf("Expected second entry ID entry-2, got %s", results[1].ID)
	}
}

func TestGetHabitEntriesHandler_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoWithFindByID{
		errorToReturn: errors.ErrNotFound,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "non-existent",
		UserID:  "user-123",
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestGetHabitEntriesHandler_ReturnsErrorWhenUserDoesNotOwnHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-456", // Different user
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestGetHabitEntriesHandler_FiltersDeletedEntries(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	date1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

	entry1 := entities.NewHabitEntry("habit-1", date1, nil)
	entry1.ID = "entry-1"

	entry2 := entities.NewHabitEntry("habit-1", date2, nil)
	entry2.ID = "entry-2"
	now := time.Now()
	entry2.DeletedAt = &now // This one is deleted

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: []*entities.HabitEntry{entry1, entry2},
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only return the non-deleted entry
	if len(results) != 1 {
		t.Fatalf("Expected 1 active entry, got %d", len(results))
	}

	if results[0].ID != "entry-1" {
		t.Errorf("Expected entry-1 (non-deleted), got %s", results[0].ID)
	}
}
