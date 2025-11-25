package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"habit-tracker-api/internal/domain/entities"
	"habit-tracker-api/internal/shared/errors"

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
			specific_days, specific_dates, carry_over, target_value, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		habit.TargetValue,
		habit.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create habit: %w", err)
	}

	return nil
}

func (r *HabitRepository) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, target_value,
			   created_at, archived_at
		FROM habits
		WHERE id = ?
	`

	var (
		habit         entities.Habit
		specificDays  sql.NullString
		specificDates sql.NullString
		archivedAt    sql.NullTime
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
		&habit.TargetValue,
		&habit.CreatedAt,
		&archivedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find habit: %w", err)
	}

	if specificDays.Valid {
		json.Unmarshal([]byte(specificDays.String), &habit.SpecificDays)
	}
	if specificDates.Valid {
		json.Unmarshal([]byte(specificDates.String), &habit.SpecificDates)
	}
	if archivedAt.Valid {
		habit.ArchivedAt = &archivedAt.Time
	}

	return &habit, nil
}

func (r *HabitRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	query := `
		SELECT id, user_id, name, description, type, frequency,
			   specific_days, specific_dates, carry_over, target_value,
			   created_at, archived_at
		FROM habits
		WHERE user_id = ? AND archived_at IS NULL
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
	specificDays, _ := json.Marshal(habit.SpecificDays)
	specificDates, _ := json.Marshal(habit.SpecificDates)

	query := `
		UPDATE habits
		SET name = ?, description = ?, type = ?, frequency = ?,
			specific_days = ?, specific_dates = ?, carry_over = ?,
			target_value = ?, archived_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		habit.Name,
		habit.Description,
		habit.Type,
		habit.Frequency,
		specificDays,
		specificDates,
		habit.CarryOver,
		habit.TargetValue,
		habit.ArchivedAt,
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
			archivedAt    sql.NullTime
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
			&habit.TargetValue,
			&habit.CreatedAt,
			&archivedAt,
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
		if archivedAt.Valid {
			habit.ArchivedAt = &archivedAt.Time
		}

		habits = append(habits, &habit)
	}

	return habits, nil
}
