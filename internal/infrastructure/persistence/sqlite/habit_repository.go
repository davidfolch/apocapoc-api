package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
	"apocapoc-api/internal/shared/pagination"

	"github.com/google/uuid"
)

type HabitRepository struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) Create(ctx context.Context, habit *entities.Habit) error {
	habit.ID = uuid.New().String()

	specificDays, _ := json.Marshal(habit.SpecificDays)
	specificDates, _ := json.Marshal(habit.SpecificDates)

	query := `
		INSERT INTO habits (
			id, user_id, name, description, type, frequency,
			specific_days, specific_dates, carry_over, is_negative, target_value, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		habit.ID,
		habit.UserID,
		habit.Name,
		habit.Description,
		habit.Type,
		habit.Frequency,
		specificDays,
		specificDates,
		habit.CarryOver,
		habit.IsNegative,
		habit.TargetValue,
		habit.CreatedAt,
		habit.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create habit: %w", err)
	}

	return nil
}

func (r *HabitRepository) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE id = ? AND deleted_at IS NULL
	`

	var (
		habit         entities.Habit
		specificDays  sql.NullString
		specificDates sql.NullString
		updatedAt     sql.NullTime
		archivedAt    sql.NullTime
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&habit.ID,
		&habit.UserID,
		&habit.Name,
		&habit.Description,
		&habit.Type,
		&habit.Frequency,
		&specificDays,
		&specificDates,
		&habit.CarryOver,
		&habit.IsNegative,
		&habit.TargetValue,
		&habit.CreatedAt,
		&updatedAt,
		&archivedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find habit: %w", err)
	}

	if updatedAt.Valid {
		habit.UpdatedAt = updatedAt.Time
	}
	if archivedAt.Valid {
		habit.ArchivedAt = &archivedAt.Time
	}
	if deletedAt.Valid {
		habit.DeletedAt = &deletedAt.Time
	}

	if specificDays.Valid {
		json.Unmarshal([]byte(specificDays.String), &habit.SpecificDays)
	}
	if specificDates.Valid {
		json.Unmarshal([]byte(specificDates.String), &habit.SpecificDates)
	}

	return &habit, nil
}

func (r *HabitRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE user_id = ? AND archived_at IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find habits: %w", err)
	}
	defer rows.Close()

	return r.scanHabits(rows)
}

func (r *HabitRepository) Update(ctx context.Context, habit *entities.Habit) error {
	habit.Touch()

	specificDays, _ := json.Marshal(habit.SpecificDays)
	specificDates, _ := json.Marshal(habit.SpecificDates)

	query := `
		UPDATE habits
		SET name = ?, description = ?, type = ?, frequency = ?,
			specific_days = ?, specific_dates = ?, carry_over = ?, is_negative = ?,
			target_value = ?, archived_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		habit.Name,
		habit.Description,
		habit.Type,
		habit.Frequency,
		specificDays,
		specificDates,
		habit.CarryOver,
		habit.IsNegative,
		habit.TargetValue,
		habit.ArchivedAt,
		habit.UpdatedAt,
		habit.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update habit: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *HabitRepository) scanHabits(rows *sql.Rows) ([]*entities.Habit, error) {
	var habits []*entities.Habit

	for rows.Next() {
		var (
			habit         entities.Habit
			specificDays  sql.NullString
			specificDates sql.NullString
			updatedAt     sql.NullTime
			archivedAt    sql.NullTime
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.Type,
			&habit.Frequency,
			&specificDays,
			&specificDates,
			&habit.CarryOver,
			&habit.IsNegative,
			&habit.TargetValue,
			&habit.CreatedAt,
			&updatedAt,
			&archivedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, err
		}

		if specificDays.Valid {
			json.Unmarshal([]byte(specificDays.String), &habit.SpecificDays)
		}
		if specificDates.Valid {
			json.Unmarshal([]byte(specificDates.String), &habit.SpecificDates)
		}
		if updatedAt.Valid {
			habit.UpdatedAt = updatedAt.Time
		}
		if archivedAt.Valid {
			habit.ArchivedAt = &archivedAt.Time
		}
		if deletedAt.Valid {
			habit.DeletedAt = &deletedAt.Time
		}

		habits = append(habits, &habit)
	}

	return habits, nil
}

func (r *HabitRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find habits: %w", err)
	}
	defer rows.Close()

	return r.scanHabits(rows)
}

func (r *HabitRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM habits WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete habit: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *HabitRepository) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE user_id = ? AND archived_at IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, params.Limit(), params.Offset())
	if err != nil {
		return nil, fmt.Errorf("failed to find habits: %w", err)
	}
	defer rows.Close()

	return r.scanHabits(rows)
}

func (r *HabitRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM habits
		WHERE user_id = ? AND archived_at IS NULL AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count habits: %w", err)
	}

	return count, nil
}

func (r *HabitRepository) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	baseQuery := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE user_id = ?`

	args := []interface{}{userID}
	conditions := []string{}

	// Always exclude soft deleted
	conditions = append(conditions, "deleted_at IS NULL")

	if !filter.IncludeArchived {
		conditions = append(conditions, "archived_at IS NULL")
	}

	if filter.Type != nil {
		conditions = append(conditions, "type = ?")
		args = append(args, string(*filter.Type))
	}

	if filter.Frequency != nil {
		conditions = append(conditions, "frequency = ?")
		args = append(args, string(*filter.Frequency))
	}

	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	for _, condition := range conditions {
		baseQuery += " AND " + condition
	}

	baseQuery += " ORDER BY created_at DESC"

	if paginationParams != nil {
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, paginationParams.Limit(), paginationParams.Offset())
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find habits: %w", err)
	}
	defer rows.Close()

	return r.scanHabits(rows)
}

func (r *HabitRepository) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	baseQuery := `SELECT COUNT(*) FROM habits WHERE user_id = ?`

	args := []interface{}{userID}
	conditions := []string{}

	// Always exclude soft deleted
	conditions = append(conditions, "deleted_at IS NULL")

	if !filter.IncludeArchived {
		conditions = append(conditions, "archived_at IS NULL")
	}

	if filter.Type != nil {
		conditions = append(conditions, "type = ?")
		args = append(args, string(*filter.Type))
	}

	if filter.Frequency != nil {
		conditions = append(conditions, "frequency = ?")
		args = append(args, string(*filter.Frequency))
	}

	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	for _, condition := range conditions {
		baseQuery += " AND " + condition
	}

	var count int
	err := r.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count habits: %w", err)
	}

	return count, nil
}

func (r *HabitRepository) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	changes := &repositories.HabitChanges{
		Created: []*entities.Habit{},
		Updated: []*entities.Habit{},
		Deleted: []string{},
	}

	// Get created and updated habits (not deleted)
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, is_negative, target_value,
			   created_at, updated_at, archived_at, deleted_at
		FROM habits
		WHERE user_id = ?
		  AND updated_at > ?
		  AND deleted_at IS NULL
		ORDER BY updated_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query habits changes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			habit         entities.Habit
			specificDays  sql.NullString
			specificDates sql.NullString
			updatedAt     sql.NullTime
			archivedAt    sql.NullTime
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.Type,
			&habit.Frequency,
			&specificDays,
			&specificDates,
			&habit.CarryOver,
			&habit.IsNegative,
			&habit.TargetValue,
			&habit.CreatedAt,
			&updatedAt,
			&archivedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}

		if specificDays.Valid {
			json.Unmarshal([]byte(specificDays.String), &habit.SpecificDays)
		}
		if specificDates.Valid {
			json.Unmarshal([]byte(specificDates.String), &habit.SpecificDates)
		}
		if updatedAt.Valid {
			habit.UpdatedAt = updatedAt.Time
		}
		if archivedAt.Valid {
			habit.ArchivedAt = &archivedAt.Time
		}
		if deletedAt.Valid {
			habit.DeletedAt = &deletedAt.Time
		}

		// Classify as created or updated based on when it was created
		if habit.CreatedAt.After(since) {
			changes.Created = append(changes.Created, &habit)
		} else {
			changes.Updated = append(changes.Updated, &habit)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habits: %w", err)
	}

	// Get deleted habits
	queryDeleted := `
		SELECT id
		FROM habits
		WHERE user_id = ?
		  AND deleted_at IS NOT NULL
		  AND deleted_at > ?
		ORDER BY deleted_at ASC
	`

	rowsDeleted, err := r.db.QueryContext(ctx, queryDeleted, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query deleted habits: %w", err)
	}
	defer rowsDeleted.Close()

	for rowsDeleted.Next() {
		var id string
		if err := rowsDeleted.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan deleted habit id: %w", err)
		}
		changes.Deleted = append(changes.Deleted, id)
	}

	if err := rowsDeleted.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deleted habits: %w", err)
	}

	return changes, nil
}

func (r *HabitRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()

	query := `
		UPDATE habits
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete habit: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}
