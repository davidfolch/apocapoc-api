package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/i18n"
	"apocapoc-api/internal/infrastructure/auth"
	"apocapoc-api/internal/infrastructure/backup"
	"apocapoc-api/internal/infrastructure/config"
	"apocapoc-api/internal/infrastructure/crypto"
	"apocapoc-api/internal/infrastructure/email"
	httpInfra "apocapoc-api/internal/infrastructure/http"
	"apocapoc-api/internal/infrastructure/logger"
	"apocapoc-api/internal/infrastructure/persistence/sqlite"
)

// @title Apocapoc API
// @description Self-hosted habit tracking service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/davidfolch/apocapoc-api
// @contact.email contact@apocapoc.app

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init(logger.Config{
		Level:       cfg.LogLevel,
		Environment: cfg.Environment,
	})

	db, err := sqlite.NewDatabase(cfg.DBPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	backupInterval, err := parseDuration(cfg.BackupInterval)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid BACKUP_INTERVAL")
	}

	backupRetentionDays, err := strconv.Atoi(cfg.BackupRetentionDays)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid BACKUP_RETENTION_DAYS")
	}

	backupScheduler := backup.NewScheduler(db.Conn(), backup.Config{
		Enabled:       cfg.BackupEnabled == "true",
		Interval:      backupInterval,
		RetentionDays: backupRetentionDays,
		Path:          cfg.BackupPath,
		Compress:      cfg.BackupCompress == "true",
		DatabasePath:  cfg.DBPath,
	})
	backupScheduler.Start()
	defer backupScheduler.Stop()

	jwtExpiryHours, err := parseJWTExpiry(cfg.JWTExpiry)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid JWT_EXPIRY")
	}

	refreshTokenExpiry, err := parseDuration(cfg.RefreshTokenExpiry)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid REFRESH_TOKEN_EXPIRY")
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, jwtExpiryHours)
	passwordHasher := crypto.NewBcryptHasher()

	var emailService *email.SMTPService
	if cfg.SMTPHost != "" {
		smtpPort, err := strconv.Atoi(cfg.SMTPPort)
		if err != nil {
			logger.Fatal().Err(err).Msg("Invalid SMTP_PORT")
		}

		emailService = email.NewSMTPService(email.SMTPConfig{
			Host:         cfg.SMTPHost,
			Port:         smtpPort,
			Username:     cfg.SMTPUser,
			Password:     cfg.SMTPPassword,
			From:         cfg.SMTPFrom,
			SupportEmail: cfg.SupportEmail,
		})
	}

	sendWelcomeEmail := cfg.SendWelcomeEmail == "true"

	userRepo := sqlite.NewUserRepository(db.Conn())
	habitRepo := sqlite.NewHabitRepository(db.Conn())
	entryRepo := sqlite.NewHabitEntryRepository(db.Conn())
	refreshTokenRepo := sqlite.NewRefreshTokenRepository(db.Conn())
	passwordResetTokenRepo := sqlite.NewPasswordResetTokenRepository(db.Conn())

	translator, err := i18n.NewTranslator()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create translator")
	}

	registerHandler := commands.NewRegisterUserHandler(userRepo, passwordHasher, emailService, cfg.AppURL, cfg.RegistrationMode, sendWelcomeEmail)
	loginHandler := queries.NewLoginUserHandler(userRepo, passwordHasher)
	refreshTokenHandler := queries.NewRefreshTokenHandler(refreshTokenRepo, userRepo)
	revokeTokenHandler := commands.NewRevokeTokenHandler(refreshTokenRepo)
	revokeAllTokensHandler := commands.NewRevokeAllTokensHandler(refreshTokenRepo)
	verifyEmailHandler := commands.NewVerifyEmailHandler(userRepo, emailService, sendWelcomeEmail)
	resendVerificationEmailHandler := commands.NewResendVerificationEmailHandler(userRepo, emailService, cfg.AppURL)
	requestPasswordResetHandler := commands.NewRequestPasswordResetHandler(userRepo, passwordResetTokenRepo, emailService, cfg.AppURL)
	resetPasswordHandler := commands.NewResetPasswordHandler(userRepo, passwordResetTokenRepo, passwordHasher)
	deleteUserHandler := commands.NewDeleteUserHandler(userRepo)
	createHandler := commands.NewCreateHabitHandler(habitRepo)
	getTodaysHandler := queries.NewGetTodaysHabitsHandler(habitRepo, entryRepo)
	getUserHabitsHandler := queries.NewGetUserHabitsHandler(habitRepo)
	getHabitByIDHandler := queries.NewGetHabitByIDHandler(habitRepo)
	getHabitEntriesHandler := queries.NewGetHabitEntriesHandler(habitRepo, entryRepo)
	getHabitStatsHandler := queries.NewGetHabitStatsHandler(habitRepo, entryRepo)
	exportUserDataHandler := queries.NewExportUserDataHandler(habitRepo, entryRepo)
	updateHandler := commands.NewUpdateHabitHandler(habitRepo)
	archiveHandler := commands.NewArchiveHabitHandler(habitRepo)
	markHandler := commands.NewMarkHabitHandler(entryRepo, habitRepo)
	unmarkHandler := commands.NewUnmarkHabitHandler(habitRepo, entryRepo)
	getSyncChangesHandler := queries.NewGetSyncChangesHandler(habitRepo, entryRepo)
	applySyncBatchHandler := commands.NewApplySyncBatchHandler(habitRepo, entryRepo)

	authHandlers := httpInfra.NewAuthHandlers(registerHandler, loginHandler, refreshTokenHandler, revokeTokenHandler, revokeAllTokensHandler, verifyEmailHandler, resendVerificationEmailHandler, requestPasswordResetHandler, resetPasswordHandler, jwtService, refreshTokenRepo, refreshTokenExpiry, translator)
	habitHandlers := httpInfra.NewHabitHandlers(createHandler, getTodaysHandler, getUserHabitsHandler, getHabitByIDHandler, getHabitEntriesHandler, updateHandler, archiveHandler, markHandler, unmarkHandler, translator)
	statsHandlers := httpInfra.NewStatsHandlers(getHabitStatsHandler, translator)
	healthHandlers := httpInfra.NewHealthHandlers(db.Conn(), emailService)
	userHandlers := httpInfra.NewUserHandlers(deleteUserHandler, translator)
	exportHandlers := httpInfra.NewExportHandlers(exportUserDataHandler, translator)
	syncHandlers := httpInfra.NewSyncHandlers(getSyncChangesHandler, applySyncBatchHandler, translator)

	router := httpInfra.NewRouter(cfg.AppURL, habitHandlers, authHandlers, statsHandlers, healthHandlers, userHandlers, exportHandlers, syncHandlers, jwtService, translator)

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	logger.Info().Str("address", addr).Msg("Server starting")

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}

func parseJWTExpiry(expiry string) (int, error) {
	expiry = strings.TrimSpace(expiry)
	if strings.HasSuffix(expiry, "h") {
		hours := strings.TrimSuffix(expiry, "h")
		return strconv.Atoi(hours)
	}
	return 0, fmt.Errorf("invalid format, expected format like '24h'")
}

func parseDuration(duration string) (time.Duration, error) {
	duration = strings.TrimSpace(duration)
	if strings.HasSuffix(duration, "m") {
		minutes := strings.TrimSuffix(duration, "m")
		mins, err := strconv.Atoi(minutes)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes value: %w", err)
		}
		return time.Duration(mins) * time.Minute, nil
	}
	if strings.HasSuffix(duration, "h") {
		hours := strings.TrimSuffix(duration, "h")
		hrs, err := strconv.Atoi(hours)
		if err != nil {
			return 0, fmt.Errorf("invalid hours value: %w", err)
		}
		return time.Duration(hrs) * time.Hour, nil
	}
	if strings.HasSuffix(duration, "d") {
		days := strings.TrimSuffix(duration, "d")
		dys, err := strconv.Atoi(days)
		if err != nil {
			return 0, fmt.Errorf("invalid days value: %w", err)
		}
		return time.Duration(dys) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("invalid format, expected format like '15m', '24h', or '7d'")
}
