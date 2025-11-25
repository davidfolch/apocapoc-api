package sqlite

import (
	"context"
	"testing"
	"time"

	"habit-tracker-api/internal/domain/entities"
	"habit-tracker-api/internal/domain/value_objects"
	"habit-tracker-api/internal/shared/errors"
)

func TestHabitRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-123",
		"Morning Exercise",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
	)
	habit.Description = "Exercise every morning"

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if habit.ID == "" {
		t.Error("Expected habit ID to be generated, got empty string")
	}
}

func TestHabitRepositoryCreateWithSpecificDays(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-123",
		"Weekly Workout",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyWeekly,
		false,
	)
	habit.SpecificDays = []int{1, 3, 5}

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, habit.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if len(found.SpecificDays) != 3 {
		t.Errorf("Expected 3 specific days, got %d", len(found.SpecificDays))
	}
}

func TestHabitRepositoryFindByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-456",
		"Read Books",
		value_objects.HabitTypeCounter,
		value_objects.FrequencyDaily,
		true,
	)
	targetValue := 30.0
	habit.TargetValue = &targetValue

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, habit.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.ID != habit.ID {
		t.Errorf("Expected ID %s, got %s", habit.ID, found.ID)
	}
	if found.Name != habit.Name {
		t.Errorf("Expected name %s, got %s", habit.Name, found.Name)
	}
	if found.CarryOver != habit.CarryOver {
		t.Errorf("Expected carry_over %v, got %v", habit.CarryOver, found.CarryOver)
	}
	if found.TargetValue == nil || *found.TargetValue != 30.0 {
		t.Errorf("Expected target_value 30.0, got %v", found.TargetValue)
	}
}

func TestHabitRepositoryFindByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "non-existent-id")
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestHabitRepositoryFindActiveByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	userID := "user-789"

	habit1 := entities.NewHabit(userID, "Habit 1", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit2 := entities.NewHabit(userID, "Habit 2", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)
	habit3 := entities.NewHabit(userID, "Habit 3", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false)

	repo.Create(ctx, habit1)
	repo.Create(ctx, habit2)
	repo.Create(ctx, habit3)

	now := time.Now()
	habit2.ArchivedAt = &now
	repo.Update(ctx, habit2)

	habits, err := repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindActiveByUserID failed: %v", err)
	}

	if len(habits) != 2 {
		t.Errorf("Expected 2 active habits, got %d", len(habits))
	}
}

func TestHabitRepositoryUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-999",
		"Original Name",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	habit.Name = "Updated Name"
	habit.Description = "Updated description"
	habit.CarryOver = true

	err = repo.Update(ctx, habit)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.FindByID(ctx, habit.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", found.Name)
	}
	if found.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got %s", found.Description)
	}
	if !found.CarryOver {
		t.Error("Expected carry_over to be true")
	}
}

func TestHabitRepositoryUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-123",
		"Test",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
	)
	habit.ID = "non-existent"

	err := repo.Update(ctx, habit)
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestHabitRepositoryArchive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	habit := entities.NewHabit(
		"user-archive",
		"To Archive",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	now := time.Now()
	habit.ArchivedAt = &now

	err = repo.Update(ctx, habit)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.FindByID(ctx, habit.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.ArchivedAt == nil {
		t.Error("Expected habit to be archived")
	}
}
