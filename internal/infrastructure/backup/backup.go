package backup

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"apocapoc-api/internal/infrastructure/logger"
)

type Config struct {
	Enabled       bool
	Interval      time.Duration
	RetentionDays int
	Path          string
	Compress      bool
	DatabasePath  string
}

func CreateBackup(db *sql.DB, config Config) error {
	if !config.Enabled {
		return nil
	}

	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("apocapoc_%s.db", timestamp)
	backupPath := filepath.Join(config.Path, filename)

	logger.Info().
		Str("backup_path", backupPath).
		Msg("Starting database backup")

	if err := backupDatabase(db, backupPath); err != nil {
		logger.Error().
			Err(err).
			Str("backup_path", backupPath).
			Msg("Backup failed")
		return fmt.Errorf("backup failed: %w", err)
	}

	if config.Compress {
		compressedPath := backupPath + ".gz"
		if err := compressFile(backupPath, compressedPath); err != nil {
			logger.Warn().
				Err(err).
				Str("backup_path", backupPath).
				Msg("Compression failed, keeping uncompressed backup")
		} else {
			os.Remove(backupPath)
			backupPath = compressedPath
		}
	}

	logger.Info().
		Str("backup_path", backupPath).
		Msg("Backup completed successfully")

	return nil
}

func backupDatabase(db *sql.DB, destPath string) error {
	_, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", destPath))
	if err != nil {
		return fmt.Errorf("vacuum into failed: %w", err)
	}
	return nil
}

func compressFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}
