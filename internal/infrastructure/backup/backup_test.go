package backup

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestCreateBackup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES ('test1'), ('test2')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	backupPath := filepath.Join(tempDir, "backups")
	config := Config{
		Enabled:       true,
		Interval:      24 * time.Hour,
		RetentionDays: 7,
		Path:          backupPath,
		Compress:      false,
		DatabasePath:  dbPath,
	}

	err = CreateBackup(db, config)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	files, err := os.ReadDir(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 backup file, got %d", len(files))
	}

	if len(files) > 0 && filepath.Ext(files[0].Name()) != ".db" {
		t.Errorf("Expected backup file to have .db extension, got %s", files[0].Name())
	}
}

func TestCreateBackupWithCompression(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	backupPath := filepath.Join(tempDir, "backups")
	config := Config{
		Enabled:       true,
		Interval:      24 * time.Hour,
		RetentionDays: 7,
		Path:          backupPath,
		Compress:      true,
		DatabasePath:  dbPath,
	}

	err = CreateBackup(db, config)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	files, err := os.ReadDir(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 backup file, got %d", len(files))
	}

	if len(files) > 0 && filepath.Ext(files[0].Name()) != ".gz" {
		t.Errorf("Expected backup file to have .gz extension, got %s", files[0].Name())
	}
}

func TestCreateBackupDisabled(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupPath := filepath.Join(tempDir, "backups")
	config := Config{
		Enabled:       false,
		Interval:      24 * time.Hour,
		RetentionDays: 7,
		Path:          backupPath,
		Compress:      false,
		DatabasePath:  dbPath,
	}

	err = CreateBackup(db, config)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	_, err = os.Stat(backupPath)
	if !os.IsNotExist(err) {
		t.Error("Backup directory should not exist when backup is disabled")
	}
}

func TestCleanOldBackups(t *testing.T) {
	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "backups")
	os.MkdirAll(backupPath, 0755)

	oldFile := filepath.Join(backupPath, "apocapoc_20200101_120000.db")
	recentFile := filepath.Join(backupPath, "apocapoc_"+time.Now().Format("20060102_150405")+".db")

	os.WriteFile(oldFile, []byte("old"), 0644)
	os.WriteFile(recentFile, []byte("recent"), 0644)

	oldTime := time.Now().AddDate(0, 0, -10)
	os.Chtimes(oldFile, oldTime, oldTime)

	config := Config{
		Enabled:       true,
		RetentionDays: 7,
		Path:          backupPath,
	}

	err := CleanOldBackups(config)
	if err != nil {
		t.Fatalf("CleanOldBackups failed: %v", err)
	}

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old backup file should have been deleted")
	}

	if _, err := os.Stat(recentFile); err != nil {
		t.Error("Recent backup file should still exist")
	}
}
