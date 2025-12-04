package backup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"apocapoc-api/internal/infrastructure/logger"
)

func CleanOldBackups(config Config) error {
	if !config.Enabled {
		return nil
	}

	if config.RetentionDays <= 0 {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -config.RetentionDays)

	files, err := os.ReadDir(config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	deletedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "apocapoc_") {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".db") && !strings.HasSuffix(file.Name(), ".db.gz") {
			continue
		}

		filePath := filepath.Join(config.Path, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("file", filePath).
				Msg("Failed to stat backup file")
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				logger.Warn().
					Err(err).
					Str("file", filePath).
					Msg("Failed to delete old backup")
				continue
			}

			logger.Info().
				Str("file", file.Name()).
				Time("mod_time", info.ModTime()).
				Msg("Deleted old backup")
			deletedCount++
		}
	}

	if deletedCount > 0 {
		logger.Info().
			Int("deleted_count", deletedCount).
			Int("retention_days", config.RetentionDays).
			Msg("Backup cleanup completed")
	}

	return nil
}
