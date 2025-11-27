package sqlite

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

func TestHabitEntryRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	scheduledDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	value := 5.0
	entry := entities.NewHabitEntry("habit-123", scheduledDate, &value)

	err := repo.Create(ctx, entry)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if entry.ID == "" {
		t.Error("Expected entry ID to be generated, got empty string")
	}
}

func TestHabitEntryRepositoryCreateDuplicateDate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	scheduledDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	habitID := "habit-456"

	entry1 := entities.NewHabitEntry(habitID, scheduledDate, nil)
	err := repo.Create(ctx, entry1)
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	entry2 := entities.NewHabitEntry(habitID, scheduledDate, nil)
	err = repo.Create(ctx, entry2)
	if err != errors.ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists, got %v", err)
	}
}

func TestHabitEntryRepositoryFindByHabitIDAndDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	habitID := "habit-range-test"

	date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	date4 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	entry1 := entities.NewHabitEntry(habitID, date1, nil)
	entry2 := entities.NewHabitEntry(habitID, date2, nil)
	entry3 := entities.NewHabitEntry(habitID, date3, nil)
	entry4 := entities.NewHabitEntry(habitID, date4, nil)

	repo.Create(ctx, entry1)
	repo.Create(ctx, entry2)
	repo.Create(ctx, entry3)
	repo.Create(ctx, entry4)

	from := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)

	entries, err := repo.FindByHabitIDAndDateRange(ctx, habitID, from, to)
	if err != nil {
		t.Fatalf("FindByHabitIDAndDateRange failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in range, got %d", len(entries))
	}
}

func TestHabitEntryRepositoryUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	scheduledDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	entry := entities.NewHabitEntry("habit-update", scheduledDate, nil)

	err := repo.Create(ctx, entry)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := repo.FindByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if retrieved.ID != entry.ID {
		t.Fatalf("Expected entry ID %s, got %s", entry.ID, retrieved.ID)
	}
}

func TestHabitEntryRepositoryUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	scheduledDate := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	entry := entities.NewHabitEntry("habit-123", scheduledDate, nil)
	entry.ID = "non-existent"

	err := repo.Update(ctx, entry)
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestHabitEntryRepositoryWithValue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	scheduledDate := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	value := 42.5
	entry := entities.NewHabitEntry("habit-value", scheduledDate, &value)

	err := repo.Create(ctx, entry)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	from := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)

	entries, err := repo.FindByHabitIDAndDateRange(ctx, "habit-value", from, to)
	if err != nil {
		t.Fatalf("FindByHabitIDAndDateRange failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Value == nil || *entries[0].Value != 42.5 {
		t.Errorf("Expected value 42.5, got %v", entries[0].Value)
	}
}

func TestHabitEntryRepositoryOrderedByDate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitEntryRepository(db)
	ctx := context.Background()

	habitID := "habit-order"

	date3 := time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)
	date1 := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 5, 10, 0, 0, 0, 0, time.UTC)

	repo.Create(ctx, entities.NewHabitEntry(habitID, date3, nil))
	repo.Create(ctx, entities.NewHabitEntry(habitID, date1, nil))
	repo.Create(ctx, entities.NewHabitEntry(habitID, date2, nil))

	from := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC)

	entries, err := repo.FindByHabitIDAndDateRange(ctx, habitID, from, to)
	if err != nil {
		t.Fatalf("FindByHabitIDAndDateRange failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	if entries[0].ScheduledDate.Format("2006-01-02") != date1.Format("2006-01-02") {
		t.Errorf("First entry should be %s, got %s", date1.Format("2006-01-02"), entries[0].ScheduledDate.Format("2006-01-02"))
	}
	if entries[1].ScheduledDate.Format("2006-01-02") != date2.Format("2006-01-02") {
		t.Errorf("Second entry should be %s, got %s", date2.Format("2006-01-02"), entries[1].ScheduledDate.Format("2006-01-02"))
	}
	if entries[2].ScheduledDate.Format("2006-01-02") != date3.Format("2006-01-02") {
		t.Errorf("Third entry should be %s, got %s", date3.Format("2006-01-02"), entries[2].ScheduledDate.Format("2006-01-02"))
	}
}
