package commands

import (
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type mockHabitRepoForUpdate struct {
	mockHabitRepo
	habitToReturn *entities.Habit
	errorOnFind   error
	errorOnUpdate error
	updatedHabit  *entities.Habit
}

func (m *mockHabitRepoForUpdate) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	if m.errorOnFind != nil {
		return nil, m.errorOnFind
	}
	return m.habitToReturn, nil
}

func (m *mockHabitRepoForUpdate) Update(ctx context.Context, habit *entities.Habit) error {
	if m.errorOnUpdate != nil {
		return m.errorOnUpdate
	}
	m.updatedHabit = habit
	return nil
}

func TestUpdateHabitHandler_UpdatesSuccessfully(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewUpdateHabitHandler(habitRepo)

	newTargetValue := 5.0
	cmd := UpdateHabitCommand{
		HabitID:      "habit-1",
		UserID:       "user-123",
		Name:         "Morning Exercise",
		Description:  "Updated description",
		CarryOver:    true,
		TargetValue:  &newTargetValue,
		SpecificDays: []int{1, 3, 5},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitRepo.updatedHabit.Name != "Morning Exercise" {
		t.Errorf("Expected name to be updated to 'Morning Exercise', got %s", habitRepo.updatedHabit.Name)
	}

	if habitRepo.updatedHabit.Description != "Updated description" {
		t.Errorf("Expected description to be updated, got %s", habitRepo.updatedHabit.Description)
	}

	if !habitRepo.updatedHabit.CarryOver {
		t.Error("Expected CarryOver to be true")
	}

	if habitRepo.updatedHabit.TargetValue == nil || *habitRepo.updatedHabit.TargetValue != 5.0 {
		t.Errorf("Expected target value 5.0, got %v", habitRepo.updatedHabit.TargetValue)
	}

	if len(habitRepo.updatedHabit.SpecificDays) != 3 {
		t.Errorf("Expected 3 specific days, got %d", len(habitRepo.updatedHabit.SpecificDays))
	}
}

func TestUpdateHabitHandler_ReturnsErrorWhenHabitNotFound(t *testing.T) {
	habitRepo := &mockHabitRepoForUpdate{
		errorOnFind: errors.ErrNotFound,
	}

	handler := NewUpdateHabitHandler(habitRepo)

	cmd := UpdateHabitCommand{
		HabitID: "non-existent",
		UserID:  "user-123",
		Name:    "Exercise",
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpdateHabitHandler_ReturnsErrorWhenUserDoesNotOwnHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewUpdateHabitHandler(habitRepo)

	cmd := UpdateHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-456", // Different user
		Name:    "Exercise",
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrUnauthorized {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestUpdateHabitHandler_CannotUpdateArchivedHabit(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"
	habit.Archive()

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewUpdateHabitHandler(habitRepo)

	cmd := UpdateHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-123",
		Name:    "Updated Exercise",
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput for archived habit, got %v", err)
	}
}

func TestUpdateHabitHandler_ValidatesInput(t *testing.T) {
	habit := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit.ID = "habit-1"

	habitRepo := &mockHabitRepoForUpdate{
		habitToReturn: habit,
	}

	handler := NewUpdateHabitHandler(habitRepo)

	cmd := UpdateHabitCommand{
		HabitID: "habit-1",
		UserID:  "user-123",
		Name:    "", // Empty name
	}

	err := handler.Handle(context.Background(), cmd)

	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput for empty name, got %v", err)
	}
}
