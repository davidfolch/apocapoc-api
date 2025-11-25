package sqlite

import (
	"os"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	dbPath := "./test_db.sqlite"
	defer os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}
}

func TestNewDatabaseCreatesDataDirectory(t *testing.T) {
	dbPath := "./test_data/nested/db.sqlite"
	defer os.RemoveAll("./test_data")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat("./test_data/nested"); os.IsNotExist(err) {
		t.Fatal("Data directory was not created")
	}
}

func TestNewDatabaseRunsMigrations(t *testing.T) {
	dbPath := ":memory:"

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	conn := db.Conn()
	var name string
	err = conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	if err != nil {
		t.Fatal("Migrations were not run, users table does not exist")
	}
}

func TestDatabaseClose(t *testing.T) {
	dbPath := ":memory:"

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestDatabaseConn(t *testing.T) {
	dbPath := ":memory:"

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	conn := db.Conn()
	if conn == nil {
		t.Fatal("Expected sql.DB connection, got nil")
	}

	err = conn.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}
