package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
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
		INSERT INTO habit_entries (id, habit_id, scheduled_date, completed_at, value, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.HabitID,
		entry.ScheduledDate.Format("2006-01-02"),
		entry.CompletedAt,
		entry.Value,
		entry.UpdatedAt,
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
		SELECT id, habit_id, scheduled_date, completed_at, value, updated_at, deleted_at
		FROM habit_entries
		WHERE habit_id = ?
		  AND scheduled_date >= ?
		  AND scheduled_date <= ?
		  AND deleted_at IS NULL
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
	entry.UpdatedAt = time.Now()

	query := `
		UPDATE habit_entries
		SET value = ?, completed_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, entry.Value, entry.CompletedAt, entry.UpdatedAt, entry.ID)
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
			updatedAt     sql.NullTime
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&scheduledDate,
			&entry.CompletedAt,
			&entry.Value,
			&updatedAt,
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

		if updatedAt.Valid {
			entry.UpdatedAt = updatedAt.Time
		}
		if deletedAt.Valid {
			entry.DeletedAt = &deletedAt.Time
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *HabitEntryRepository) FindByID(ctx context.Context, id string) (*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value, updated_at, deleted_at
		FROM habit_entries
		WHERE id = ? AND deleted_at IS NULL
	`

	var (
		entry         entities.HabitEntry
		scheduledDate string
		updatedAt     sql.NullTime
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.HabitID,
		&scheduledDate,
		&entry.CompletedAt,
		&entry.Value,
		&updatedAt,
		&deletedAt,
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

	if updatedAt.Valid {
		entry.UpdatedAt = updatedAt.Time
	}
	if deletedAt.Valid {
		entry.DeletedAt = &deletedAt.Time
	}

	return &entry, nil
}

func (r *HabitEntryRepository) FindByHabitID(ctx context.Context, habitID string) ([]*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value, updated_at, deleted_at
		FROM habit_entries
		WHERE habit_id = ? AND deleted_at IS NULL
		ORDER BY scheduled_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *HabitEntryRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.HabitEntry, error) {
	query := `
		SELECT he.id, he.habit_id, he.scheduled_date, he.completed_at, he.value, he.updated_at, he.deleted_at
		FROM habit_entries he
		INNER JOIN habits h ON he.habit_id = h.id
		WHERE h.user_id = ? AND he.deleted_at IS NULL
		ORDER BY he.scheduled_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

func (r *HabitEntryRepository) FindPendingByHabitID(ctx context.Context, habitID string, beforeDate time.Time) ([]*entities.HabitEntry, error) {
	query := `
		SELECT id, habit_id, scheduled_date, completed_at, value, updated_at, deleted_at
		FROM habit_entries
		WHERE habit_id = ?
		  AND scheduled_date < ?
		  AND deleted_at IS NULL
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

func (r *HabitEntryRepository) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitEntryChanges, error) {
	changes := &repositories.HabitEntryChanges{
		Created: []*entities.HabitEntry{},
		Updated: []*entities.HabitEntry{},
		Deleted: []string{},
	}

	query := `
		SELECT he.id, he.habit_id, he.scheduled_date, he.completed_at, he.value, he.updated_at, he.deleted_at
		FROM habit_entries he
		INNER JOIN habits h ON he.habit_id = h.id
		WHERE h.user_id = ?
		  AND he.updated_at > ?
		  AND he.deleted_at IS NULL
		ORDER BY he.updated_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query habit entry changes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			entry         entities.HabitEntry
			scheduledDate string
			updatedAt     sql.NullTime
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&scheduledDate,
			&entry.CompletedAt,
			&entry.Value,
			&updatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan habit entry: %w", err)
		}

		parsedDate, err := time.Parse("2006-01-02", scheduledDate)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, scheduledDate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse scheduled_date: %w", err)
			}
		}
		entry.ScheduledDate = parsedDate

		if updatedAt.Valid {
			entry.UpdatedAt = updatedAt.Time
		}
		if deletedAt.Valid {
			entry.DeletedAt = &deletedAt.Time
		}

		if entry.CompletedAt.After(since) {
			changes.Created = append(changes.Created, &entry)
		} else {
			changes.Updated = append(changes.Updated, &entry)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habit entries: %w", err)
	}

	queryDeleted := `
		SELECT he.id
		FROM habit_entries he
		INNER JOIN habits h ON he.habit_id = h.id
		WHERE h.user_id = ?
		  AND he.deleted_at IS NOT NULL
		  AND he.deleted_at > ?
		ORDER BY he.deleted_at ASC
	`

	rowsDeleted, err := r.db.QueryContext(ctx, queryDeleted, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query deleted habit entries: %w", err)
	}
	defer rowsDeleted.Close()

	for rowsDeleted.Next() {
		var id string
		if err := rowsDeleted.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan deleted habit entry id: %w", err)
		}
		changes.Deleted = append(changes.Deleted, id)
	}

	if err := rowsDeleted.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deleted habit entries: %w", err)
	}

	return changes, nil
}

func (r *HabitEntryRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()

	query := `
		UPDATE habit_entries
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete habit entry: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.ErrNotFound
	}

	return nil
}
