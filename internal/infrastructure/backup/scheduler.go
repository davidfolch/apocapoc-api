package backup

import (
	"database/sql"
	"time"

	"apocapoc-api/internal/infrastructure/logger"
)

type Scheduler struct {
	db     *sql.DB
	config Config
	stopCh chan struct{}
}

func NewScheduler(db *sql.DB, config Config) *Scheduler {
	return &Scheduler{
		db:     db,
		config: config,
		stopCh: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	if !s.config.Enabled {
		logger.Info().Msg("Backup scheduler is disabled")
		return
	}

	logger.Info().
		Dur("interval", s.config.Interval).
		Int("retention_days", s.config.RetentionDays).
		Str("path", s.config.Path).
		Bool("compress", s.config.Compress).
		Msg("Starting backup scheduler")

	go s.run()
}

func (s *Scheduler) run() {
	if err := CreateBackup(s.db, s.config); err != nil {
		logger.Error().Err(err).Msg("Initial backup failed")
	}

	if err := CleanOldBackups(s.config); err != nil {
		logger.Error().Err(err).Msg("Initial cleanup failed")
	}

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Debug().Msg("Running scheduled backup")

			if err := CreateBackup(s.db, s.config); err != nil {
				logger.Error().Err(err).Msg("Scheduled backup failed")
				continue
			}

			if err := CleanOldBackups(s.config); err != nil {
				logger.Error().Err(err).Msg("Backup cleanup failed")
			}

		case <-s.stopCh:
			logger.Info().Msg("Backup scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	if s.config.Enabled {
		close(s.stopCh)
	}
}
