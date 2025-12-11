package commands

import (
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type mockEntryRepo struct {
	createFunc          func(ctx context.Context, entry *entities.HabitEntry) error
	findByDateRangeFunc func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error)
	updateFunc          func(ctx context.Context, entry *entities.HabitEntry) error
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
	if m.findByDateRangeFunc != nil {
		return m.findByDateRangeFunc(ctx, habitID, from, to)
	}
	return nil, nil
}

func (m *mockEntryRepo) FindByUserID(ctx context.Context, userID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepo) Update(ctx context.Context, entry *entities.HabitEntry) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, entry)
	}
	return nil
}

func (m *mockEntryRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockEntryRepo) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitEntryChanges, error) {
	return &repositories.HabitEntryChanges{
		Created: []*entities.HabitEntry{},
		Updated: []*entities.HabitEntry{},
		Deleted: []string{},
	}, nil
}

func (m *mockEntryRepo) SoftDelete(ctx context.Context, id string) error {
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
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
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
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
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
	habit := entities.NewHabit("user-123", "Steps", value_objects.HabitTypeValue, value_objects.FrequencyDaily, false, false)
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
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
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

func TestMarkHabitHandler_CounterOnlyAcceptsIntegers(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	decimalValue := 2.5
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &decimalValue,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput for decimal value on COUNTER, got %v", err)
	}
}

func TestMarkHabitHandler_CounterAcceptsIntegers(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 3.0 {
				t.Errorf("Expected value 3.0, got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	intValue := 3.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &intValue,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_CounterAutoIncrementFirstMark(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{}, nil
		},
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 1.0 {
				t.Errorf("Expected default value 1.0, got %v", entry.Value)
			}
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

func TestMarkHabitHandler_CounterAutoIncrementSubsequentMarks(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	existingValue := 3.0
	existingEntry := &entities.HabitEntry{
		ID:            "entry-1",
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &existingValue,
	}

	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{existingEntry}, nil
		},
		updateFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 4.0 {
				t.Errorf("Expected value 4.0, got %v", entry.Value)
			}
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

func TestMarkHabitHandler_CounterAutoIncrementWithCustomValue(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	existingValue := 3.0
	existingEntry := &entities.HabitEntry{
		ID:            "entry-1",
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &existingValue,
	}

	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{existingEntry}, nil
		},
		updateFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 5.0 {
				t.Errorf("Expected value 5.0 (3+2), got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	increment := 2.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &increment,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_CounterCanDecrement(t *testing.T) {
	habit := entities.NewHabit("user-123", "Cigarettes", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	existingValue := 5.0
	existingEntry := &entities.HabitEntry{
		ID:            "entry-1",
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &existingValue,
	}

	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{existingEntry}, nil
		},
		updateFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 3.0 {
				t.Errorf("Expected value 3.0 (5-2), got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	decrement := -2.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &decrement,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_CounterMinimumZero(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	existingValue := 2.0
	existingEntry := &entities.HabitEntry{
		ID:            "entry-1",
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &existingValue,
	}

	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{existingEntry}, nil
		},
		updateFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 0.0 {
				t.Errorf("Expected value 0.0 (2-3 clamped to 0), got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	decrement := -3.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &decrement,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_CounterStaysAtZero(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}

	existingValue := 0.0
	existingEntry := &entities.HabitEntry{
		ID:            "entry-1",
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &existingValue,
	}

	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{existingEntry}, nil
		},
		updateFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 0.0 {
				t.Errorf("Expected value to stay at 0.0, got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	decrement := -1.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &decrement,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMarkHabitHandler_CounterFirstMarkWithNegative(t *testing.T) {
	habit := entities.NewHabit("user-123", "Water Glasses", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForMark{habit: habit}
	entryRepo := &mockEntryRepo{
		findByDateRangeFunc: func(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
			return []*entities.HabitEntry{}, nil
		},
		createFunc: func(ctx context.Context, entry *entities.HabitEntry) error {
			if entry.Value == nil || *entry.Value != 0.0 {
				t.Errorf("Expected value 0.0 (negative clamped), got %v", entry.Value)
			}
			return nil
		},
	}

	handler := NewMarkHabitHandler(entryRepo, habitRepo)

	negativeValue := -5.0
	cmd := MarkHabitCommand{
		HabitID:       "habit-1",
		ScheduledDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Value:         &negativeValue,
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func (m *mockEntryRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockEntryRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForMark) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForMark) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockEntryRepo) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockEntryRepo) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForMark) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForMark) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForMark) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	return &repositories.HabitChanges{
		Created: []*entities.Habit{},
		Updated: []*entities.Habit{},
		Deleted: []string{},
	}, nil
}

func (m *mockHabitRepoForMark) SoftDelete(ctx context.Context, id string) error {
	return nil
}
