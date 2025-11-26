package queries

import (
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
)

func TestGetUserHabitsHandler_ReturnsAllActiveHabits(t *testing.T) {
	habit1 := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit1.ID = "habit-1"

	habit2 := entities.NewHabit("user-123", "Read", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false)
	habit2.ID = "habit-2"

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit1, habit2}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-123",
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 habits, got %d", len(results))
	}

	if results[0].ID != "habit-1" {
		t.Errorf("Expected first habit ID habit-1, got %s", results[0].ID)
	}

	if results[1].ID != "habit-2" {
		t.Errorf("Expected second habit ID habit-2, got %s", results[1].ID)
	}
}

func TestGetUserHabitsHandler_ReturnsEmptyListForUserWithNoHabits(t *testing.T) {
	habitRepo := &mockHabitRepo{habits: []*entities.Habit{}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-456",
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Expected 0 habits, got %d", len(results))
	}
}

func TestGetUserHabitsHandler_IncludesAllHabitFields(t *testing.T) {
	targetValue := 5.0
	habit := entities.NewHabit("user-123", "Drink Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, true)
	habit.ID = "habit-1"
	habit.TargetValue = &targetValue

	habitRepo := &mockHabitRepo{habits: []*entities.Habit{habit}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-123",
	}

	results, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(results))
	}

	result := results[0]

	if result.Name != "Drink Water" {
		t.Errorf("Expected name 'Drink Water', got %s", result.Name)
	}

	if result.Type != value_objects.HabitTypeValue {
		t.Errorf("Expected type %s, got %s", value_objects.HabitTypeValue, result.Type)
	}

	if result.Frequency != value_objects.FrequencyDaily {
		t.Errorf("Expected frequency %s, got %s", value_objects.FrequencyDaily, result.Frequency)
	}

	if result.TargetValue == nil || *result.TargetValue != 5.0 {
		t.Errorf("Expected target value 5.0, got %v", result.TargetValue)
	}

	if !result.CarryOver {
		t.Error("Expected carry over to be true")
	}
}
