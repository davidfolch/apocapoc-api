package sqlite

import (
	"database/sql"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createHabitsTable,
		createHabitEntriesTable,
		createRefreshTokensTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}
	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	timezone TEXT DEFAULT 'UTC',
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

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_habits_user ON habits(user_id);
CREATE INDEX IF NOT EXISTS idx_habits_active ON habits(user_id, archived_at);
CREATE INDEX IF NOT EXISTS idx_entries_habit ON habit_entries(habit_id);
CREATE INDEX IF NOT EXISTS idx_entries_scheduled ON habit_entries(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
`
