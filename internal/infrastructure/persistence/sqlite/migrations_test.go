package sqlite

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db)

	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	tables := []string{"users", "habits", "habit_entries"}
	for _, table := range tables {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		err := db.QueryRow(query, table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}
}

func TestUsersTableSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	columns := []string{"id", "email", "password_hash", "timezone", "created_at", "updated_at"}
	for _, col := range columns {
		query := "SELECT " + col + " FROM users LIMIT 0"
		rows, err := db.Query(query)
		if err != nil {
			t.Errorf("Column %s does not exist in users table: %v", col, err)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

func TestHabitsTableSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	columns := []string{"id", "user_id", "name", "description", "type", "frequency",
		"specific_days", "specific_dates", "carry_over", "target_value", "created_at", "archived_at"}
	for _, col := range columns {
		query := "SELECT " + col + " FROM habits LIMIT 0"
		rows, err := db.Query(query)
		if err != nil {
			t.Errorf("Column %s does not exist in habits table: %v", col, err)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

func TestHabitEntriesTableSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	columns := []string{"id", "habit_id", "scheduled_date", "completed_at", "value"}
	for _, col := range columns {
		query := "SELECT " + col + " FROM habit_entries LIMIT 0"
		rows, err := db.Query(query)
		if err != nil {
			t.Errorf("Column %s does not exist in habit_entries table: %v", col, err)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("First RunMigrations failed: %v", err)
	}

	err = RunMigrations(db)

	if err != nil {
		t.Errorf("Second RunMigrations should be idempotent but failed: %v", err)
	}
}
