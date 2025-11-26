package queries

import (
	"context"
	"strconv"
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

func (m *mockEntryRepoWithFindByHabitID) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
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
		Page:    1,
		Limit:   50,
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(result.Entries))
	}

	if result.Entries[0].ID != "entry-1" {
		t.Errorf("Expected first entry ID entry-1, got %s", result.Entries[0].ID)
	}

	if result.Entries[1].ID != "entry-2" {
		t.Errorf("Expected second entry ID entry-2, got %s", result.Entries[1].ID)
	}

	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
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
		Page:    1,
		Limit:   50,
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
		UserID:  "user-456",
		Page:    1,
		Limit:   50,
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestGetHabitEntriesHandler_WithPagination(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	entries := make([]*entities.HabitEntry, 10)
	for i := 0; i < 10; i++ {
		date := time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC)
		entry := entities.NewHabitEntry("habit-1", date, nil)
		entry.ID = "entry-" + strconv.Itoa(i+1)
		entries[i] = entry
	}

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: entries,
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
		Page:    2,
		Limit:   3,
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Entries) != 3 {
		t.Fatalf("Expected 3 entries on page 2, got %d", len(result.Entries))
	}

	if result.Total != 10 {
		t.Errorf("Expected total 10, got %d", result.Total)
	}

	if result.Page != 2 {
		t.Errorf("Expected page 2, got %d", result.Page)
	}

	if result.Limit != 3 {
		t.Errorf("Expected limit 3, got %d", result.Limit)
	}
}

func TestGetHabitEntriesHandler_RequiresPaginationWithoutDateRange(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: []*entities.HabitEntry{},
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
		Page:    0,
		Limit:   0,
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput when pagination not provided without date range, got %v", err)
	}
}

func TestGetHabitEntriesHandler_AllowsNoPaginationWithShortDateRange(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)

	entry := entities.NewHabitEntry("habit-1", from, nil)
	entry.ID = "entry-1"

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: []*entities.HabitEntry{entry},
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
		From:    &from,
		To:      &to,
		Page:    0,
		Limit:   0,
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error for short date range without pagination, got %v", err)
	}

	if len(result.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result.Entries))
	}
}

func TestGetHabitEntriesHandler_RequiresPaginationWithLongDateRange(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	entryRepo := &mockEntryRepoWithFindByHabitID{
		entries: []*entities.HabitEntry{},
	}

	handler := NewGetHabitEntriesHandler(habitRepo, entryRepo)

	query := GetHabitEntriesQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
		From:    &from,
		To:      &to,
		Page:    0,
		Limit:   0,
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput for date range > 1 year without pagination, got %v", err)
	}
}

