package sqlite

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
)

func TestHabitRepository_GetChangesSince_EmptyWhenNoChanges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Crear hábito inicial
	habit := entities.NewHabit(
		userID,
		"Initial Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Timestamp después de la creación
	time.Sleep(10 * time.Millisecond)
	since := time.Now()

	// No hay cambios después de 'since'
	changes, err := repo.GetChangesSince(ctx, userID, since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	if len(changes.Created) != 0 {
		t.Errorf("Expected 0 created habits, got %d", len(changes.Created))
	}
	if len(changes.Updated) != 0 {
		t.Errorf("Expected 0 updated habits, got %d", len(changes.Updated))
	}
	if len(changes.Deleted) != 0 {
		t.Errorf("Expected 0 deleted habits, got %d", len(changes.Deleted))
	}
}

func TestHabitRepository_GetChangesSince_ReturnsCreatedHabits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Timestamp de referencia
	since := time.Now()
	time.Sleep(10 * time.Millisecond)

	// Crear hábito DESPUÉS de 'since'
	habit := entities.NewHabit(
		userID,
		"New Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Obtener cambios
	changes, err := repo.GetChangesSince(ctx, userID, since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	if len(changes.Created) != 1 {
		t.Fatalf("Expected 1 created habit, got %d", len(changes.Created))
	}

	if changes.Created[0].Name != "New Habit" {
		t.Errorf("Expected habit name 'New Habit', got '%s'", changes.Created[0].Name)
	}

	if changes.Created[0].ID != habit.ID {
		t.Errorf("Expected habit ID '%s', got '%s'", habit.ID, changes.Created[0].ID)
	}
}

func TestHabitRepository_GetChangesSince_ReturnsUpdatedHabits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Crear hábito inicial
	habit := entities.NewHabit(
		userID,
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

	// Timestamp de referencia
	time.Sleep(10 * time.Millisecond)
	since := time.Now()
	time.Sleep(10 * time.Millisecond)

	// Actualizar hábito DESPUÉS de 'since'
	habit.Name = "Updated Name"
	err = repo.Update(ctx, habit)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Obtener cambios
	changes, err := repo.GetChangesSince(ctx, userID, since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	if len(changes.Updated) != 1 {
		t.Fatalf("Expected 1 updated habit, got %d", len(changes.Updated))
	}

	if changes.Updated[0].Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got '%s'", changes.Updated[0].Name)
	}

	if len(changes.Created) != 0 {
		t.Errorf("Expected 0 created habits (should be in Updated), got %d", len(changes.Created))
	}
}

func TestHabitRepository_GetChangesSince_ReturnsDeletedHabits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Crear hábito
	habit := entities.NewHabit(
		userID,
		"To Delete",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	habitID := habit.ID

	// Timestamp de referencia
	time.Sleep(10 * time.Millisecond)
	since := time.Now()
	time.Sleep(10 * time.Millisecond)

	// Soft delete DESPUÉS de 'since'
	err = repo.SoftDelete(ctx, habitID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// Obtener cambios
	changes, err := repo.GetChangesSince(ctx, userID, since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	if len(changes.Deleted) != 1 {
		t.Fatalf("Expected 1 deleted habit, got %d", len(changes.Deleted))
	}

	if changes.Deleted[0] != habitID {
		t.Errorf("Expected deleted habit ID '%s', got '%s'", habitID, changes.Deleted[0])
	}
}

func TestHabitRepository_GetChangesSince_CombinedChanges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Crear hábito inicial (antes de 'since')
	habitOld := entities.NewHabit(
		userID,
		"Old Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habitOld)

	// Timestamp de referencia
	time.Sleep(10 * time.Millisecond)
	since := time.Now()
	time.Sleep(10 * time.Millisecond)

	// DESPUÉS de 'since':
	// 1. Crear nuevo hábito
	habitNew := entities.NewHabit(
		userID,
		"New Habit",
		value_objects.HabitTypeCounter,
		value_objects.FrequencyWeekly,
		false,
		false,
	)
	habitNew.SpecificDays = []int{1, 3, 5}
	repo.Create(ctx, habitNew)

	// 2. Actualizar hábito existente
	habitOld.Name = "Old Habit Updated"
	repo.Update(ctx, habitOld)

	// 3. Crear y eliminar otro hábito
	habitToDelete := entities.NewHabit(
		userID,
		"To Delete",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habitToDelete)
	repo.SoftDelete(ctx, habitToDelete.ID)

	// Obtener cambios
	changes, err := repo.GetChangesSince(ctx, userID, since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	// Verificar creados (habitNew, NO habitToDelete porque fue eliminado)
	if len(changes.Created) != 1 {
		t.Errorf("Expected 1 created habit, got %d", len(changes.Created))
	}
	if len(changes.Created) > 0 && changes.Created[0].Name != "New Habit" {
		t.Errorf("Expected created habit name 'New Habit', got '%s'", changes.Created[0].Name)
	}

	// Verificar actualizados
	if len(changes.Updated) != 1 {
		t.Errorf("Expected 1 updated habit, got %d", len(changes.Updated))
	}
	if len(changes.Updated) > 0 && changes.Updated[0].Name != "Old Habit Updated" {
		t.Errorf("Expected updated habit name 'Old Habit Updated', got '%s'", changes.Updated[0].Name)
	}

	// Verificar eliminados
	if len(changes.Deleted) != 1 {
		t.Errorf("Expected 1 deleted habit, got %d", len(changes.Deleted))
	}
	if len(changes.Deleted) > 0 && changes.Deleted[0] != habitToDelete.ID {
		t.Errorf("Expected deleted habit ID '%s', got '%s'", habitToDelete.ID, changes.Deleted[0])
	}
}

func TestHabitRepository_GetChangesSince_OnlyReturnsUserHabits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	since := time.Now()
	time.Sleep(10 * time.Millisecond)

	// Crear hábitos de diferentes usuarios
	habitUser1 := entities.NewHabit(
		"user-1",
		"User 1 Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habitUser1)

	habitUser2 := entities.NewHabit(
		"user-2",
		"User 2 Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habitUser2)

	// Obtener cambios solo de user-1
	changes, err := repo.GetChangesSince(ctx, "user-1", since)
	if err != nil {
		t.Fatalf("GetChangesSince failed: %v", err)
	}

	if len(changes.Created) != 1 {
		t.Fatalf("Expected 1 created habit for user-1, got %d", len(changes.Created))
	}

	if changes.Created[0].UserID != "user-1" {
		t.Errorf("Expected user ID 'user-1', got '%s'", changes.Created[0].UserID)
	}
}

func TestHabitRepository_SoftDelete_MarksAsDeleted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Crear hábito
	habit := entities.NewHabit(
		"user-123",
		"To Delete",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habit)

	// Soft delete
	err := repo.SoftDelete(ctx, habit.ID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// El hábito NO debe aparecer en FindByID (porque está eliminado)
	found, err := repo.FindByID(ctx, habit.ID)
	if err == nil {
		t.Error("Expected error when finding soft-deleted habit, got nil")
	}
	if found != nil {
		t.Error("Expected nil habit when soft-deleted, got habit")
	}
}

func TestHabitRepository_SoftDelete_NotFoundError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Intentar eliminar hábito inexistente
	err := repo.SoftDelete(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent habit, got nil")
	}
}

func TestHabitRepository_SoftDelete_CannotDeleteTwice(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Crear hábito
	habit := entities.NewHabit(
		"user-123",
		"To Delete",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habit)

	// Primera eliminación
	err := repo.SoftDelete(ctx, habit.ID)
	if err != nil {
		t.Fatalf("First SoftDelete failed: %v", err)
	}

	// Segunda eliminación debe fallar
	err = repo.SoftDelete(ctx, habit.ID)
	if err == nil {
		t.Error("Expected error when deleting already deleted habit, got nil")
	}
}

func TestHabitRepository_Update_UpdatesUpdatedAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Crear hábito
	habit := entities.NewHabit(
		"user-123",
		"Original",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habit)

	originalUpdatedAt := habit.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	// Actualizar
	habit.Name = "Updated"
	err := repo.Update(ctx, habit)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verificar que UpdatedAt cambió
	if !habit.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("Expected UpdatedAt to be updated, but it wasn't. Original: %v, Current: %v",
			originalUpdatedAt, habit.UpdatedAt)
	}
}

func TestHabitRepository_Create_SetsUpdatedAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()

	// Crear hábito
	habit := entities.NewHabit(
		"user-123",
		"New Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)

	err := repo.Create(ctx, habit)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verificar que UpdatedAt está seteado
	if habit.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set, got zero value")
	}

	// UpdatedAt debe ser igual a CreatedAt al crear
	if !habit.UpdatedAt.Equal(habit.CreatedAt) {
		t.Errorf("Expected UpdatedAt to equal CreatedAt on creation. UpdatedAt: %v, CreatedAt: %v",
			habit.UpdatedAt, habit.CreatedAt)
	}
}

func TestHabitRepository_FindActiveByUserID_ExcludesSoftDeleted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewHabitRepository(db)
	ctx := context.Background()
	userID := "user-123"

	// Crear 2 hábitos
	habit1 := entities.NewHabit(
		userID,
		"Active Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habit1)

	habit2 := entities.NewHabit(
		userID,
		"Deleted Habit",
		value_objects.HabitTypeBoolean,
		value_objects.FrequencyDaily,
		false,
		false,
	)
	repo.Create(ctx, habit2)

	// Soft delete uno
	repo.SoftDelete(ctx, habit2.ID)

	// FindActiveByUserID debe devolver solo el activo
	activeHabits, err := repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindActiveByUserID failed: %v", err)
	}

	if len(activeHabits) != 1 {
		t.Fatalf("Expected 1 active habit, got %d", len(activeHabits))
	}

	if activeHabits[0].Name != "Active Habit" {
		t.Errorf("Expected 'Active Habit', got '%s'", activeHabits[0].Name)
	}
}
