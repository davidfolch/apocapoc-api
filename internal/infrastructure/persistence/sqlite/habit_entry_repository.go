package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"

	"github.com/google/uuid"
)

type HabitEntryRepository struct {
	db *sql.DB
}

func NewHabitEntryRepository(db *sql.DB) *HabitEntryRepository {
	return &HabitEntryRepository{db: db}
}

func (r *HabitEntryRepository) Create(ctx context.Context, entry *entities.HabitEntry) error {
	entry.ID = uuid.New().String()

	query := `
		INSERT INTO habit_entries (id, habit_id, scheduled_date, completed_at, value)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.HabitID,
		entry.ScheduledDate.Format("2006-01-02"),
		entry.CompletedAt,
		entry.Value,
	)

	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

func (r *HabitEntryRepository) FindByHabitIDAndDateRange(
	ctx context.Context,
	habitID string,
	from, to time.Time,
) ([]*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value
		FROM habit_entries
		WHERE habit_id = ?
		  AND scheduled_date >= ?
		  AND scheduled_date <= ?
		ORDER BY scheduled_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query,
		habitID,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *HabitEntryRepository) Update(ctx context.Context, entry *entities.HabitEntry) error {
	query := `
		UPDATE habit_entries
		SET value = ?, completed_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, entry.Value, entry.CompletedAt, entry.ID)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *HabitEntryRepository) scanEntries(rows *sql.Rows) ([]*entities.HabitEntry, error) {
	var entries []*entities.HabitEntry

	for rows.Next() {
		var (
			entry         entities.HabitEntry
			scheduledDate string
		)

		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&scheduledDate,
			&entry.CompletedAt,
			&entry.Value,
		)

		if err != nil {
			return nil, err
		}

		parsedDate, err := time.Parse("2006-01-02", scheduledDate)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, scheduledDate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse scheduled_date: %w", err)
			}
		}
		entry.ScheduledDate = parsedDate

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *HabitEntryRepository) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value
		FROM habit_entries
		WHERE id = ?
	`

	var (
		entry         entities.HabitEntry
		scheduledDate string
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.HabitID,
		&scheduledDate,
		&entry.CompletedAt,
		&entry.Value,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find entry: %w", err)
	}

	parsedDate, err := time.Parse("2006-01-02", scheduledDate)
	if err != nil {
		parsedDate, err = time.Parse(time.RFC3339, scheduledDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse scheduled_date: %w", err)
		}
	}
	entry.ScheduledDate = parsedDate

	return &entry, nil
}

func (r *HabitEntryRepository) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value
		FROM habit_entries
		WHERE habit_id = ?
		ORDER BY scheduled_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *HabitEntryRepository) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value
		FROM habit_entries
		WHERE habit_id = ?
		  AND scheduled_date < ?
		ORDER BY scheduled_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, habitID, beforeDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to find pending entries: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *HabitEntryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM habit_entries WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}
