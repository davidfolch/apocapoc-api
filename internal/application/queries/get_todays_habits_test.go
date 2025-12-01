package queries

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/pagination"
)

type mockHabitRepo struct {
	habits []*entities.Habit
}

func (m *mockHabitRepo) Create(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepo) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return m.habits, nil
}

func (m *mockHabitRepo) Update(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockHabitRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

type mockEntryRepo struct {
	entries []*entities.HabitEntry
}

func (m *mockEntryRepo) Create(ctx context.Context, entry *entities.HabitEntry) error {
	return nil
}

func (m *mockEntryRepo) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
	var result []*entities.HabitEntry
	for _, e := range m.entries {
		if e.HabitID == habitID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockEntryRepo) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) Update(ctx context.Context, entry *entities.HabitEntry) error {
	return nil
}

func (m *mockEntryRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func TestGetTodaysHabitsHandler_DailyHabitNoEntries(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(results))
	}

	if results[0].ID != "habit-1" {
		t.Errorf("Expected habit ID habit-1, got %s", results[0].ID)
	}

	if results[0].IsCarriedOver {
		t.Error("Expected IsCarriedOver to be false")
	}

	if results[0].Entry != nil {
		t.Error("Expected entry to be nil when no entry exists")
	}
}

func TestGetTodaysHabitsHandler_DailyHabitAlreadyCompleted(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	targetDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	entry := entities.NewHabitEntry("habit-1", targetDate, nil)
	entry.ID = "entry-1"

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{entry}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     targetDate,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit (with entry), got %d", len(results))
	}

	if results[0].Entry == nil {
		t.Fatal("Expected entry to be present")
	}

	if results[0].Entry.ID != "entry-1" {
		t.Errorf("Expected entry ID entry-1, got %s", results[0].Entry.ID)
	}
}

func TestGetTodaysHabitsHandler_HabitWithValueEntry(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"
	targetValue := 2000.0
	habit.TargetValue = &targetValue

	targetDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	value := 1500.0
	entry := entities.NewHabitEntry("habit-1", targetDate, &value)
	entry.ID = "entry-1"

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{entry}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     targetDate,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(results))
	}

	if results[0].Entry == nil {
		t.Fatal("Expected entry to be present")
	}

	if results[0].Entry.Value == nil {
		t.Fatal("Expected entry value to be present")
	}

	if *results[0].Entry.Value != 1500.0 {
		t.Errorf("Expected entry value 1500.0, got %f", *results[0].Entry.Value)
	}
}

func TestGetTodaysHabitsHandler_WeeklyHabitOnCorrectDay(t *testing.T) {
	habit := entities.NewHabit("user-123", "Gym", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	habit.ID = "habit-1"
	habit.SpecificDays = []int{1, 3, 5}

	monday := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     monday,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(results))
	}
}

func TestGetTodaysHabitsHandler_WeeklyHabitOnWrongDay(t *testing.T) {
	habit := entities.NewHabit("user-123", "Gym", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	habit.ID = "habit-1"
	habit.SpecificDays = []int{1, 3, 5}

	tuesday := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     tuesday,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Expected 0 habits (wrong day), got %d", len(results))
	}
}

func TestGetTodaysHabitsHandler_CarryOverEnabled(t *testing.T) {
	habit := entities.NewHabit("user-123", "Gym", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, true, false)
	habit.ID = "habit-1"
	habit.SpecificDays = []int{1}

	tuesday := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     tuesday,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit (carry-over), got %d", len(results))
	}

	if !results[0].IsCarriedOver {
		t.Error("Expected IsCarriedOver to be true")
	}
}

func TestGetTodaysHabitsHandler_CarryOverDisabled(t *testing.T) {
	habit := entities.NewHabit("user-123", "Gym", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	habit.ID = "habit-1"
	habit.SpecificDays = []int{1}

	tuesday := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}
	entryRepo := &mockEntryRepo{entries: []*entities.HabitEntry{}}

	handler := NewGetTodaysHabitsHandler(habitRepo, entryRepo)

	query := GetTodaysHabitsQuery{
		UserID:   "user-123",
		Timezone: "UTC",
		Date:     tuesday,
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Expected 0 habits (no carry-over), got %d", len(results))
	}
}
