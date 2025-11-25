package entities

import (
	"testing"
	"time"
)

func TestNewHabitEntry(t *testing.T) {
	habitID := "habit-123"
	scheduledDate := time.Now().Truncate(24 * time.Hour)
	value := 5.0

	entry := NewHabitEntry(habitID, scheduledDate, &value)

	if entry.HabitID != habitID {
		t.Errorf("Expected HabitID %s, got %s", habitID, entry.HabitID)
	}

	if !entry.ScheduledDate.Equal(scheduledDate) {
		t.Errorf("Expected ScheduledDate %v, got %v", scheduledDate, entry.ScheduledDate)
	}

	if entry.Value == nil {
		t.Error("Value should not be nil")
	}

	if *entry.Value != value {
		t.Errorf("Expected Value %f, got %f", value, *entry.Value)
	}

	if entry.CompletedAt.IsZero() {
		t.Error("CompletedAt should not be zero")
	}

	if entry.DeletedAt != nil {
		t.Error("DeletedAt should be nil for new entry")
	}
}

func TestNewHabitEntry_BooleanHabit(t *testing.T) {
	habitID := "habit-123"
	scheduledDate := time.Now().Truncate(24 * time.Hour)

	entry := NewHabitEntry(habitID, scheduledDate, nil)

	if entry.Value != nil {
		t.Error("Value should be nil for boolean habit")
	}
}

func TestHabitEntry_SoftDelete(t *testing.T) {
	entry := NewHabitEntry("habit-123", time.Now(), nil)

	if entry.DeletedAt != nil {
		t.Error("New entry should not be deleted")
	}

	entry.SoftDelete()

	if entry.DeletedAt == nil {
		t.Error("Entry should be deleted after calling SoftDelete()")
	}

	if entry.DeletedAt.After(time.Now()) {
		t.Error("DeletedAt should not be in the future")
	}
}
