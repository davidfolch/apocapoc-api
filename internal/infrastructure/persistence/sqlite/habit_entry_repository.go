package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"habit-tracker-api/internal/domain/entities"
	"habit-tracker-api/internal/shared/errors"

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
		SELECT id, habit_id, scheduled_date, completed_at, value, deleted_at
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
		SET deleted_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, entry.DeletedAt, entry.ID)
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
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&scheduledDate,
			&entry.CompletedAt,
			&entry.Value,
			&deletedAt,
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

		if deletedAt.Valid {
			entry.DeletedAt = &deletedAt.Time
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}
