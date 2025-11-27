package entities

import (
	"testing"
	"time"

	"apocapoc-api/internal/domain/value_objects"
)

func TestNewHabit(t *testing.T) {
	userID := "user-123"
	name := "Morning Exercise"
	habitType := value_objects.HabitTypeBoolean
	frequency := value_objects.FrequencyDaily
	carryOver := false
	isNegative := false

	habit := NewHabit(userID, name, habitType, frequency, carryOver, isNegative)

	if habit.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, habit.UserID)
	}

	if habit.Name != name {
		t.Errorf("Expected Name %s, got %s", name, habit.Name)
	}

	if habit.Type != habitType {
		t.Errorf("Expected Type %s, got %s", habitType, habit.Type)
	}

	if habit.Frequency != frequency {
		t.Errorf("Expected Frequency %s, got %s", frequency, habit.Frequency)
	}

	if habit.CarryOver != carryOver {
		t.Errorf("Expected CarryOver %v, got %v", carryOver, habit.CarryOver)
	}

	if habit.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if habit.ArchivedAt != nil {
		t.Error("ArchivedAt should be nil for new habit")
	}
}

func TestHabit_Archive(t *testing.T) {
	habit := NewHabit("user-123", "Test Habit", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)

	if habit.ArchivedAt != nil {
		t.Error("New habit should not be archived")
	}

	habit.Archive()

	if habit.ArchivedAt == nil {
		t.Error("Habit should be archived after calling Archive()")
	}

	if habit.ArchivedAt.After(time.Now()) {
		t.Error("ArchivedAt should not be in the future")
	}
}

func TestHabit_IsActive(t *testing.T) {
	habit := NewHabit("user-123", "Test Habit", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)

	if !habit.IsActive() {
		t.Error("New habit should be active")
	}

	habit.Archive()

	if habit.IsActive() {
		t.Error("Archived habit should not be active")
	}
}

func TestHabit_WithSpecificDays(t *testing.T) {
	habit := NewHabit("user-123", "Workout", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	habit.SpecificDays = []int{1, 3, 5} // Monday, Wednesday, Friday

	if len(habit.SpecificDays) != 3 {
		t.Errorf("Expected 3 specific days, got %d", len(habit.SpecificDays))
	}

	if habit.SpecificDays[0] != 1 || habit.SpecificDays[1] != 3 || habit.SpecificDays[2] != 5 {
		t.Errorf("Expected days [1,3,5], got %v", habit.SpecificDays)
	}
}

func TestHabit_WithTargetValue(t *testing.T) {
	habit := NewHabit("user-123", "Drink Water", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	targetValue := 8.0
	habit.TargetValue = &targetValue

	if habit.TargetValue == nil {
		t.Error("TargetValue should not be nil")
	}

	if *habit.TargetValue != 8.0 {
		t.Errorf("Expected TargetValue 8.0, got %f", *habit.TargetValue)
	}
}
