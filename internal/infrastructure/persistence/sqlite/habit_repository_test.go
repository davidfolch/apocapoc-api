package sqlite

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
	"apocapoc-api/internal/shared/pagination"
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
		false,
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

	habit1 := entities.NewHabit(userID, "Habit 1", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit2 := entities.NewHabit(userID, "Habit 2", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit3 := entities.NewHabit(userID, "Habit 3", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)

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

func TestHabitRepositoryFindActiveByUserIDWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	userID := "user-pagination-test"

	for i := 1; i <= 10; i++ {
		habit := entities.NewHabit(
			userID,
			"Habit "+string(rune(i+'0')),
			value_objects.HabitTypeBoolean,
			value_objects.FrequencyDaily,
			false,
			false,
		)
		err := repo.Create(ctx, habit)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	now := time.Now()
	allHabits, _ := repo.FindActiveByUserID(ctx, userID)
	allHabits[0].ArchivedAt = &now
	repo.Update(ctx, allHabits[0])

	t.Run("FirstPage", func(t *testing.T) {
		params := pagination.NewParams(1, 5)
		habits, err := repo.FindActiveByUserIDWithPagination(ctx, userID, params)
		if err != nil {
			t.Fatalf("FindActiveByUserIDWithPagination failed: %v", err)
		}

		if len(habits) != 5 {
			t.Errorf("Expected 5 habits on first page, got %d", len(habits))
		}
	})

	t.Run("SecondPage", func(t *testing.T) {
		params := pagination.NewParams(2, 5)
		habits, err := repo.FindActiveByUserIDWithPagination(ctx, userID, params)
		if err != nil {
			t.Fatalf("FindActiveByUserIDWithPagination failed: %v", err)
		}

		if len(habits) != 4 {
			t.Errorf("Expected 4 habits on second page (9 total active), got %d", len(habits))
		}
	})

	t.Run("PageBeyondTotal", func(t *testing.T) {
		params := pagination.NewParams(10, 5)
		habits, err := repo.FindActiveByUserIDWithPagination(ctx, userID, params)
		if err != nil {
			t.Fatalf("FindActiveByUserIDWithPagination failed: %v", err)
		}

		if len(habits) != 0 {
			t.Errorf("Expected 0 habits beyond total pages, got %d", len(habits))
		}
	})

	t.Run("CustomPageSize", func(t *testing.T) {
		params := pagination.NewParams(1, 3)
		habits, err := repo.FindActiveByUserIDWithPagination(ctx, userID, params)
		if err != nil {
			t.Fatalf("FindActiveByUserIDWithPagination failed: %v", err)
		}

		if len(habits) != 3 {
			t.Errorf("Expected 3 habits with page_size=3, got %d", len(habits))
		}
	})

	t.Run("ExcludesArchived", func(t *testing.T) {
		params := pagination.NewParams(1, 20)
		habits, err := repo.FindActiveByUserIDWithPagination(ctx, userID, params)
		if err != nil {
			t.Fatalf("FindActiveByUserIDWithPagination failed: %v", err)
		}

		if len(habits) != 9 {
			t.Errorf("Expected 9 active habits (1 archived), got %d", len(habits))
		}

		for _, habit := range habits {
			if habit.ArchivedAt != nil {
				t.Error("Expected no archived habits in results")
			}
		}
	})
}

func TestHabitRepositoryCountActiveByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	userID := "user-count-test"

	t.Run("NoHabits", func(t *testing.T) {
		count, err := repo.CountActiveByUserID(ctx, "non-existent-user")
		if err != nil {
			t.Fatalf("CountActiveByUserID failed: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected count 0 for non-existent user, got %d", count)
		}
	})

	for i := 1; i <= 7; i++ {
		habit := entities.NewHabit(
			userID,
			"Habit "+string(rune(i+'0')),
			value_objects.HabitTypeBoolean,
			value_objects.FrequencyDaily,
			false,
			false,
		)
		repo.Create(ctx, habit)
	}

	t.Run("AllActive", func(t *testing.T) {
		count, err := repo.CountActiveByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("CountActiveByUserID failed: %v", err)
		}

		if count != 7 {
			t.Errorf("Expected count 7, got %d", count)
		}
	})

	t.Run("WithArchived", func(t *testing.T) {
		habits, _ := repo.FindActiveByUserID(ctx, userID)
		now := time.Now()
		habits[0].ArchivedAt = &now
		habits[1].ArchivedAt = &now
		repo.Update(ctx, habits[0])
		repo.Update(ctx, habits[1])

		count, err := repo.CountActiveByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("CountActiveByUserID failed: %v", err)
		}

		if count != 5 {
			t.Errorf("Expected count 5 (7 total - 2 archived), got %d", count)
		}
	})
}

func TestHabitRepositoryFindByUserIDFiltered(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	userID := "user-filter-test"

	habit1 := entities.NewHabit(userID, "Morning Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit1.Description = "Daily morning workout"
	repo.Create(ctx, habit1)
	time.Sleep(1 * time.Millisecond)

	habit2 := entities.NewHabit(userID, "Read Books", value_objects.HabitTypeCounter, value_objects.FrequencyWeekly, false, false)
	habit2.Description = "Read at least 3 books per week"
	repo.Create(ctx, habit2)
	time.Sleep(1 * time.Millisecond)

	habit3 := entities.NewHabit(userID, "Drink Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, false, false)
	habit3.Description = "Drink 2 liters of water daily"
	repo.Create(ctx, habit3)
	time.Sleep(1 * time.Millisecond)

	habit4 := entities.NewHabit(userID, "Weekly Run", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	repo.Create(ctx, habit4)
	time.Sleep(1 * time.Millisecond)

	now := time.Now()
	habit4.ArchivedAt = &now
	repo.Update(ctx, habit4)

	t.Run("FilterByType", func(t *testing.T) {
		habitType := value_objects.HabitTypeBoolean
		filter := repositories.HabitFilter{
			Type: &habitType,
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 1 {
			t.Errorf("Expected 1 active BOOLEAN habit, got %d", len(habits))
		}

		if habits[0].Type != value_objects.HabitTypeBoolean {
			t.Errorf("Expected BOOLEAN type, got %s", habits[0].Type)
		}
	})

	t.Run("FilterByFrequency", func(t *testing.T) {
		frequency := value_objects.FrequencyDaily
		filter := repositories.HabitFilter{
			Frequency: &frequency,
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 2 {
			t.Errorf("Expected 2 DAILY habits, got %d", len(habits))
		}
	})

	t.Run("FilterIncludeArchived", func(t *testing.T) {
		filter := repositories.HabitFilter{
			IncludeArchived: true,
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 4 {
			t.Errorf("Expected 4 habits (including archived), got %d", len(habits))
		}
	})

	t.Run("FilterBySearch", func(t *testing.T) {
		filter := repositories.HabitFilter{
			Search: "Exercise",
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 1 {
			t.Errorf("Expected 1 habit matching 'Exercise', got %d", len(habits))
		}

		if habits[0].Name != "Morning Exercise" {
			t.Errorf("Expected 'Morning Exercise', got %s", habits[0].Name)
		}
	})

	t.Run("FilterBySearchInDescription", func(t *testing.T) {
		filter := repositories.HabitFilter{
			Search: "books",
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 1 {
			t.Errorf("Expected 1 habit matching 'books' in description, got %d", len(habits))
		}
	})

	t.Run("CombineFilters", func(t *testing.T) {
		habitType := value_objects.HabitTypeBoolean
		frequency := value_objects.FrequencyWeekly
		filter := repositories.HabitFilter{
			Type:            &habitType,
			Frequency:       &frequency,
			IncludeArchived: true,
		}

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, nil)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 1 {
			t.Errorf("Expected 1 BOOLEAN WEEKLY habit (archived), got %d", len(habits))
		}

		if habits[0].Name != "Weekly Run" {
			t.Errorf("Expected 'Weekly Run', got %s", habits[0].Name)
		}
	})

	t.Run("WithPagination", func(t *testing.T) {
		filter := repositories.HabitFilter{}
		params := pagination.NewParams(1, 2)

		habits, err := repo.FindByUserIDFiltered(ctx, userID, filter, &params)
		if err != nil {
			t.Fatalf("FindByUserIDFiltered failed: %v", err)
		}

		if len(habits) != 2 {
			t.Errorf("Expected 2 habits on first page, got %d", len(habits))
		}
	})
}

func TestHabitRepositoryCountByUserIDFiltered(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	userID := "user-count-filter-test"

	habit1 := entities.NewHabit(userID, "Test1", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	repo.Create(ctx, habit1)

	habit2 := entities.NewHabit(userID, "Test2", value_objects.HabitTypeCounter, value_objects.FrequencyDaily, false, false)
	repo.Create(ctx, habit2)

	habit3 := entities.NewHabit(userID, "Test3", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	now := time.Now()
	habit3.ArchivedAt = &now
	repo.Create(ctx, habit3)
	repo.Update(ctx, habit3)

	t.Run("CountByType", func(t *testing.T) {
		habitType := value_objects.HabitTypeBoolean
		filter := repositories.HabitFilter{
			Type: &habitType,
		}

		count, err := repo.CountByUserIDFiltered(ctx, userID, filter)
		if err != nil {
			t.Fatalf("CountByUserIDFiltered failed: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 active BOOLEAN habit, got %d", count)
		}
	})

	t.Run("CountWithArchived", func(t *testing.T) {
		habitType := value_objects.HabitTypeBoolean
		filter := repositories.HabitFilter{
			Type:            &habitType,
			IncludeArchived: true,
		}

		count, err := repo.CountByUserIDFiltered(ctx, userID, filter)
		if err != nil {
			t.Fatalf("CountByUserIDFiltered failed: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 BOOLEAN habits (including archived), got %d", count)
		}
	})
}
