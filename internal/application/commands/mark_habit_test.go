package commands

import (
	"context"
	"testing"
	"time"

	"habit-tracker-api/internal/domain/entities"
	"habit-tracker-api/internal/domain/value_objects"
	"habit-tracker-api/internal/shared/errors"
)

type mockEntryRepo struct {
	createFunc func(ctx context.Context, entry *entities.HabitEntry) error
}

func (m *mockEntryRepo) Create(ctx context.Context, entry *entities.HabitEntry) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, entry)
	}
	return nil
}

func (m *mockEntryRepo) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
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

type mockHabitRepoForMark struct {
	habit *entities.Habit
}

func (m *mockHabitRepoForMark) Create(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepoForMark) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	return m.habit, nil
}

func (m *mockHabitRepoForMark) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForMark) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForMark) Update(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockHabitRepoForMark) Delete(ctx context.Context, id string) error {
	return nil
}

func TestMarkHabitHandler_Success(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.HabitID != "habit-1" {
				t.Errorf("Expected HabitID habit-1, got %s", entry.HabitID)
			}
			entry.ID = "entry-123"
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         nil,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_HabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoForMark{habit: nil}
	entryRepo := &mockEntryRepo{}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	cmd := MarkHabitCommand{
		HabitID:       "non-existent",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("Expected error when habit not found")
	}
}

func TestMarkHabitHandler_ArchivedHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"
	habit.Archive()

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("Expected error when habit is archived")
	}
}

func TestMarkHabitHandler_WithValue(t *testing.T) {
	habit := entities.NewHabit("user-123", "Steps", value_objects.HabitTypeValue, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	value := 5000.0
	entryRepo := &mockEntryRepo{
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil {
				t.Error("Expected value to be set")
			}
			if *entry.Value != value {
				t.Errorf("Expected value %f, got %f", value, *entry.Value)
			}
			entry.ID = "entry-123"
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &value,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_DuplicateEntry(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			return errors.ErrAlreadyExists
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists, got %v", err)
	}
}
