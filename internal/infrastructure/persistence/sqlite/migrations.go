package sqlite

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createHabitsTable,
		createHabitEntriesTable,
		createRefreshTokensTable,
		createPasswordResetTokensTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	if err := addEmailVerificationColumns(db); err != nil {
		return err
	}

	if err := removeTimezoneColumn(db); err != nil {
		return err
	}

	return nil
}

func addEmailVerificationColumns(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"email_verified", "ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT 0"},
		{"email_verification_token", "ALTER TABLE users ADD COLUMN email_verification_token TEXT"},
		{"email_verification_expiry", "ALTER TABLE users ADD COLUMN email_verification_expiry DATETIME"},
	}

	for _, col := range columns {
		exists, err := columnExists(db, "users", col.name)
		if err != nil {
			return err
		}

		if !exists {
			if _, err := db.Exec(col.definition); err != nil {
				return err
			}
		}
	}

	return nil
}

func removeTimezoneColumn(db *sql.DB) error {
	exists, err := columnExists(db, "users", "timezone")
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	if _, err := db.Exec("ALTER TABLE users DROP COLUMN timezone"); err != nil {
		return fmt.Errorf("failed to drop timezone column: %w", err)
	}

	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", table)
	var count int
	err := db.QueryRow(query, column).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const createHabitsTable = `
CREATE TABLE IF NOT EXISTS habits (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	type TEXT CHECK(type IN ('BOOLEAN', 'COUNTER', 'VALUE')),
	frequency TEXT CHECK(frequency IN ('DAILY', 'WEEKLY', 'MONTHLY')),
	specific_days TEXT,
	specific_dates TEXT,
	carry_over BOOLEAN DEFAULT 0,
	is_negative BOOLEAN DEFAULT 0,
	target_value REAL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	archived_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

const createHabitEntriesTable = `
CREATE TABLE IF NOT EXISTS habit_entries (
	id TEXT PRIMARY KEY,
	habit_id TEXT NOT NULL,
	scheduled_date DATE NOT NULL,
	completed_at DATETIME NOT NULL,
	value REAL,
	FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE,
	UNIQUE(habit_id, scheduled_date)
);
`

const createRefreshTokensTable = `
CREATE TABLE IF NOT EXISTS refresh_tokens (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	token TEXT UNIQUE NOT NULL,
	expires_at DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	revoked_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

const createPasswordResetTokensTable = `
CREATE TABLE IF NOT EXISTS password_reset_tokens (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	token TEXT UNIQUE NOT NULL,
	expires_at DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	used_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_habits_user ON habits(user_id);
CREATE INDEX IF NOT EXISTS idx_habits_active ON habits(user_id, archived_at);
CREATE INDEX IF NOT EXISTS idx_entries_habit ON habit_entries(habit_id);
CREATE INDEX IF NOT EXISTS idx_entries_scheduled ON habit_entries(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
`
