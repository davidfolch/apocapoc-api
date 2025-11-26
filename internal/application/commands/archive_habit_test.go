package commands

import (
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

func TestArchiveHabitHandler_ArchivesSuccessfully(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewArchiveHabitHandler(habitRepo)

	cmd := ArchiveHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-123",
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitRepo.updatedHabit.ArchivedAt == nil {
		t.Error("Expected habit to be archived")
	}

	if habitRepo.updatedHabit.IsActive() {
		t.Error("Expected habit to not be active")
	}
}

func TestArchiveHabitHandler_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoForUpdate{
		errorOnFind: errors.ErrNotFound,
	}

	handler := NewArchiveHabitHandler(habitRepo)

	cmd := ArchiveHabitCommand{
		HabitID: "non-existent",
		UserID:  "user-123",
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestArchiveHabitHandler_ReturnsErrorWhenUserDoesNotOwnHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewArchiveHabitHandler(habitRepo)

	cmd := ArchiveHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-456", // Different user
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestArchiveHabitHandler_CanArchiveAlreadyArchivedHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"
	habit.Archive()

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewArchiveHabitHandler(habitRepo)

	cmd := ArchiveHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-123",
	}

	err := handler.Handle(context.Background(), cmd)

	// Should be idempotent - no error
	if err != nil {
		t.Fatalf("Expected no error for already archived habit, got %v", err)
	}
}
