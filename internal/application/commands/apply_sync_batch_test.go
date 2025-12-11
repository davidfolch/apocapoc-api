package commands

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

type mockHabitRepoForBatch struct {
	habits       map[string]*entities.Habit
	createFunc   func(ctx context.Context, habit *entities.Habit) error
	updateFunc   func(ctx context.Context, habit *entities.Habit) error
	softDeleteFunc func(ctx context.Context, id string) error
}

func (m *mockHabitRepoForBatch) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	habit, ok := m.habits[id]
	if !ok {
		return nil, errors.ErrNotFound
	}
	return habit, nil
}

func (m *mockHabitRepoForBatch) Create(ctx context.Context, habit *entities.Habit) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, habit)
	}
	m.habits[habit.ID] = habit
	return nil
}

func (m *mockHabitRepoForBatch) Update(ctx context.Context, habit *entities.Habit) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, habit)
	}
	m.habits[habit.ID] = habit
	return nil
}

func (m *mockHabitRepoForBatch) SoftDelete(ctx context.Context, id string) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, id)
	}
	delete(m.habits, id)
	return nil
}

func (m *mockHabitRepoForBatch) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForBatch) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForBatch) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockHabitRepoForBatch) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForBatch) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockHabitRepoForBatch) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForBatch) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	return 0, nil
}

func (m *mockHabitRepoForBatch) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	return &repositories.HabitChanges{
		Created: []*entities.Habit{},
		Updated: []*entities.Habit{},
		Deleted: []string{},
	}, nil
}

type mockEntryRepoForBatch struct {
	entries        map[string]*entities.HabitEntry
	createFunc     func(ctx context.Context, entry *entities.HabitEntry) error
	updateFunc     func(ctx context.Context, entry *entities.HabitEntry) error
	softDeleteFunc func(ctx context.Context, id string) error
}

func (m *mockEntryRepoForBatch) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	entry, ok := m.entries[id]
	if !ok {
		return nil, errors.ErrNotFound
	}
	return entry, nil
}

func (m *mockEntryRepoForBatch) Create(ctx context.Context, entry *entities.HabitEntry) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, entry)
	}
	m.entries[entry.ID] = entry
	return nil
}

func (m *mockEntryRepoForBatch) Update(ctx context.Context, entry *entities.HabitEntry) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, entry)
	}
	m.entries[entry.ID] = entry
	return nil
}

func (m *mockEntryRepoForBatch) SoftDelete(ctx context.Context, id string) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, id)
	}
	delete(m.entries, id)
	return nil
}

func (m *mockEntryRepoForBatch) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForBatch) FindByHabitIDAndDateRange(ctx context.Context, habitID string, from, to time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForBatch) FindByUserID(ctx context.Context, userID string) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForBatch) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	return nil, nil
}

func (m *mockEntryRepoForBatch) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockEntryRepoForBatch) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitEntryChanges, error) {
	return &repositories.HabitEntryChanges{
		Created: []*entities.HabitEntry{},
		Updated: []*entities.HabitEntry{},
		Deleted: []string{},
	}, nil
}

func TestApplySyncBatchHandler_CreateNewHabits(t *testing.T) {
	habitRepo := &mockHabitRepoForBatch{
		habits: make(map[string]*entities.Habit),
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	newHabit := entities.NewHabit("user-123", "New Habit", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	newHabit.ID = "habit-new"

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{newHabit},
			Updated: []*entities.Habit{},
			Deleted: []string{},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(habitRepo.habits) != 1 {
		t.Errorf("Expected 1 habit created, got %d", len(habitRepo.habits))
	}
}

func TestApplySyncBatchHandler_UpdateExistingHabits(t *testing.T) {
	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	existingHabit := entities.NewHabit("user-123", "Old Name", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	existingHabit.ID = "habit-1"
	existingHabit.UpdatedAt = oldTime

	habitRepo := &mockHabitRepoForBatch{
		habits: map[string]*entities.Habit{
			"habit-1": existingHabit,
		},
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	updatedHabit := entities.NewHabit("user-123", "New Name", value_objects.HabitTypeCounter, value_objects.FrequencyWeekly, false, false)
	updatedHabit.ID = "habit-1"
	updatedHabit.UpdatedAt = newTime

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{updatedHabit},
			Deleted: []string{},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitRepo.habits["habit-1"].Name != "New Name" {
		t.Errorf("Expected habit name to be updated to 'New Name', got '%s'", habitRepo.habits["habit-1"].Name)
	}
}

func TestApplySyncBatchHandler_LastWriteWins(t *testing.T) {
	serverTime := time.Now()
	clientTime := serverTime.Add(-30 * time.Minute)

	serverHabit := entities.NewHabit("user-123", "Server Version", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	serverHabit.ID = "habit-1"
	serverHabit.UpdatedAt = serverTime

	habitRepo := &mockHabitRepoForBatch{
		habits: map[string]*entities.Habit{
			"habit-1": serverHabit,
		},
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	clientHabit := entities.NewHabit("user-123", "Client Version", value_objects.HabitTypeCounter, value_objects.FrequencyWeekly, false, false)
	clientHabit.ID = "habit-1"
	clientHabit.UpdatedAt = clientTime

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{clientHabit},
			Deleted: []string{},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if habitRepo.habits["habit-1"].Name != "Server Version" {
		t.Errorf("Expected server version to win (Last-Write-Wins), got '%s'", habitRepo.habits["habit-1"].Name)
	}
}

func TestApplySyncBatchHandler_DeleteHabits(t *testing.T) {
	existingHabit := entities.NewHabit("user-123", "To Delete", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	existingHabit.ID = "habit-1"

	habitRepo := &mockHabitRepoForBatch{
		habits: map[string]*entities.Habit{
			"habit-1": existingHabit,
		},
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{},
			Deleted: []string{"habit-1"},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(habitRepo.habits) != 0 {
		t.Errorf("Expected habit to be deleted, but still exists")
	}
}

func TestApplySyncBatchHandler_ValidatesUserOwnership(t *testing.T) {
	habitRepo := &mockHabitRepoForBatch{
		habits: make(map[string]*entities.Habit),
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	habitForDifferentUser := entities.NewHabit("user-456", "Not Yours", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habitForDifferentUser.ID = "habit-1"

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{habitForDifferentUser},
			Updated: []*entities.Habit{},
			Deleted: []string{},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("Expected error for user mismatch, got nil")
	}
}

func TestApplySyncBatchHandler_ProcessesEntries(t *testing.T) {
	habitRepo := &mockHabitRepoForBatch{
		habits: make(map[string]*entities.Habit),
	}
	entryRepo := &mockEntryRepoForBatch{
		entries: make(map[string]*entities.HabitEntry),
	}

	handler := NewApplySyncBatchHandler(habitRepo, entryRepo)

	newEntry := entities.NewHabitEntry("habit-1", time.Now(), nil)
	newEntry.ID = "entry-new"

	cmd := ApplySyncBatchCommand{
		UserID: "user-123",
		Habits: HabitBatchChanges{
			Created: []*entities.Habit{},
			Updated: []*entities.Habit{},
			Deleted: []string{},
		},
		Entries: EntryBatchChanges{
			Created: []*entities.HabitEntry{newEntry},
			Updated: []*entities.HabitEntry{},
			Deleted: []string{},
		},
	}

	err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(entryRepo.entries) != 1 {
		t.Errorf("Expected 1 entry created, got %d", len(entryRepo.entries))
	}
}
