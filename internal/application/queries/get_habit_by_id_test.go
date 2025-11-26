package queries

import (
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type mockHabitRepoWithFindByID struct {
	mockHabitRepo
	habitToReturn *entities.Habit
	errorToReturn error
}

func (m *mockHabitRepoWithFindByID) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	return m.habitToReturn, nil
}

func TestGetHabitByIDHandler_ReturnsHabitSuccessfully(t *testing.T) {
	targetValue := 5.0
	habit := entities.NewHabit("user-123", "Drink Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, true)
	habit.ID = "habit-1"
	habit.TargetValue = &targetValue

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	handler := NewGetHabitByIDHandler(habitRepo)

	query := GetHabitByIDQuery{
		HabitID: "habit-1",
		UserID:  "user-123",
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "habit-1" {
		t.Errorf("Expected ID habit-1, got %s", result.ID)
	}

	if result.Name != "Drink Water" {
		t.Errorf("Expected name 'Drink Water', got %s", result.Name)
	}

	if result.Type != value_objects.HabitTypeValue {
		t.Errorf("Expected type %s, got %s", value_objects.HabitTypeValue, result.Type)
	}

	if result.TargetValue == nil || *result.TargetValue != 5.0 {
		t.Errorf("Expected target value 5.0, got %v", result.TargetValue)
	}
}

func TestGetHabitByIDHandler_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoWithFindByID{
		errorToReturn: errors.ErrNotFound,
	}

	handler := NewGetHabitByIDHandler(habitRepo)

	query := GetHabitByIDQuery{
		HabitID: "non-existent",
		UserID:  "user-123",
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestGetHabitByIDHandler_ReturnsErrorWhenUserDoesNotOwnHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoWithFindByID{
		habitToReturn: habit,
	}

	handler := NewGetHabitByIDHandler(habitRepo)

	query := GetHabitByIDQuery{
		HabitID: "habit-1",
		UserID:  "user-456", // Different user
	}

	_, err := handler.Handle(context.Background(), query)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}
