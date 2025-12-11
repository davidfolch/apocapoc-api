package commands

import (
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockHabitRepo struct {
	createFunc func(ctx context.Context, habit *entities.Habit) error
}

func (m *mockHabitRepo) Create(ctx context.Context, habit *entities.Habit) error {
	return m.createFunc(ctx, habit)
}

func (m *mockHabitRepo) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
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

func (m *mockHabitRepo) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockHabitRepo) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	return 0, nil
}

func (m *mockHabitRepo) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	return &repositories.HabitChanges{
		Created: []*entities.Habit{},
		Updated: []*entities.Habit{},
		Deleted: []string{},
	}, nil
}

func (m *mockHabitRepo) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func TestCreateHabitHandler_Success(t *testing.T) {
	mock := &mockHabitRepo{
		createFunc: func(ctx context.Context, habit *entities.Habit) error {
			habit.ID = "habit-123"
			return nil
		},
	}

	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:      "user-123",
		Name:        "Exercise",
		Description: "Daily workout",
		Type:        "BOOLEAN",
		Frequency:   "DAILY",
		CarryOver:   true,
	}

	habitID, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitID == "" {
		t.Error("Expected habit ID to be returned")
	}
}

func TestCreateHabitHandler_InvalidType(t *testing.T) {
	mock := &mockHabitRepo{}
	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:    "user-123",
		Name:      "Exercise",
		Type:      "INVALID",
		Frequency: "DAILY",
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateHabitHandler_InvalidFrequency(t *testing.T) {
	mock := &mockHabitRepo{}
	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:    "user-123",
		Name:      "Exercise",
		Type:      "BOOLEAN",
		Frequency: "INVALID",
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateHabitHandler_WeeklyWithoutSpecificDays(t *testing.T) {
	mock := &mockHabitRepo{}
	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:       "user-123",
		Name:         "Exercise",
		Type:         "BOOLEAN",
		Frequency:    "WEEKLY",
		SpecificDays: []int{},
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateHabitHandler_MonthlyWithoutSpecificDates(t *testing.T) {
	mock := &mockHabitRepo{}
	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:        "user-123",
		Name:          "Exercise",
		Type:          "BOOLEAN",
		Frequency:     "MONTHLY",
		SpecificDates: []int{},
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateHabitHandler_WeeklyWithSpecificDays(t *testing.T) {
	mock := &mockHabitRepo{
		createFunc: func(ctx context.Context, habit *entities.Habit) error {
			habit.ID = "habit-123"
			if len(habit.SpecificDays) != 3 {
				t.Errorf("Expected 3 specific days, got %d", len(habit.SpecificDays))
			}
			return nil
		},
	}

	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:       "user-123",
		Name:         "Exercise",
		Type:         "BOOLEAN",
		Frequency:    "WEEKLY",
		SpecificDays: []int{1, 3, 5},
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCreateHabitHandler_MonthlyWithSpecificDates(t *testing.T) {
	mock := &mockHabitRepo{
		createFunc: func(ctx context.Context, habit *entities.Habit) error {
			habit.ID = "habit-123"
			if len(habit.SpecificDates) != 2 {
				t.Errorf("Expected 2 specific dates, got %d", len(habit.SpecificDates))
			}
			return nil
		},
	}

	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:        "user-123",
		Name:          "Pay bills",
		Type:          "BOOLEAN",
		Frequency:     "MONTHLY",
		SpecificDates: []int{1, 15},
	}

	_, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCreateHabitHandler_NegativeHabit(t *testing.T) {
	mock := &mockHabitRepo{
		createFunc: func(ctx context.Context, habit *entities.Habit) error {
			habit.ID = "habit-123"
			if !habit.IsNegative {
				t.Error("Expected IsNegative to be true")
			}
			return nil
		},
	}

	handler := NewCreateHabitHandler(mock)

	cmd := CreateHabitCommand{
		UserID:      "user-123",
		Name:        "Eat Candy",
		Description: "Track bad habit",
		Type:        "COUNTER",
		Frequency:   "DAILY",
		CarryOver:   false,
		IsNegative:  true,
	}

	habitID, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitID == "" {
		t.Error("Expected habit ID to be returned")
	}
}
