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
	"apocapoc-api/internal/infrastructure/auth"
	"apocapoc-api/internal/infrastructure/config"
	"apocapoc-api/internal/infrastructure/crypto"
	httpInfra "apocapoc-api/internal/infrastructure/http"
	"apocapoc-api/internal/infrastructure/persistence/sqlite"
)

// @title Apocapoc API
// @version 1.0
// @description Self-hosted habit tracking service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/davidfolch/apocapoc-api

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

	db, err := sqlite.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	jwtExpiryHours, err := parseJWTExpiry(cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("Invalid JWT_EXPIRY: %v", err)
	}

	refreshTokenExpiry, err := parseDuration(cfg.RefreshTokenExpiry)
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_EXPIRY: %v", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, jwtExpiryHours)
	passwordHasher := crypto.NewBcryptHasher()

	userRepo := sqlite.NewUserRepository(db.Conn())
	habitRepo := sqlite.NewHabitRepository(db.Conn())
	entryRepo := sqlite.NewHabitEntryRepository(db.Conn())
	refreshTokenRepo := sqlite.NewRefreshTokenRepository(db.Conn())

	registerHandler := commands.NewRegisterUserHandler(userRepo, passwordHasher)
	loginHandler := queries.NewLoginUserHandler(userRepo, passwordHasher)
	refreshTokenHandler := queries.NewRefreshTokenHandler(refreshTokenRepo, userRepo)
	revokeTokenHandler := commands.NewRevokeTokenHandler(refreshTokenRepo)
	revokeAllTokensHandler := commands.NewRevokeAllTokensHandler(refreshTokenRepo)
	createHandler := commands.NewCreateHabitHandler(habitRepo)
	getTodaysHandler := queries.NewGetTodaysHabitsHandler(habitRepo, entryRepo)
	getUserHabitsHandler := queries.NewGetUserHabitsHandler(habitRepo)
	getHabitByIDHandler := queries.NewGetHabitByIDHandler(habitRepo)
	getHabitEntriesHandler := queries.NewGetHabitEntriesHandler(habitRepo, entryRepo)
	getHabitStatsHandler := queries.NewGetHabitStatsHandler(habitRepo, entryRepo)
	updateHandler := commands.NewUpdateHabitHandler(habitRepo)
	archiveHandler := commands.NewArchiveHabitHandler(habitRepo)
	markHandler := commands.NewMarkHabitHandler(entryRepo, habitRepo)
	unmarkHandler := commands.NewUnmarkHabitHandler(habitRepo, entryRepo)

	authHandlers := httpInfra.NewAuthHandlers(registerHandler, loginHandler, refreshTokenHandler, revokeTokenHandler, revokeAllTokensHandler, jwtService, refreshTokenRepo, refreshTokenExpiry)
	habitHandlers := httpInfra.NewHabitHandlers(createHandler, getTodaysHandler, getUserHabitsHandler, getHabitByIDHandler, getHabitEntriesHandler, updateHandler, archiveHandler, markHandler, unmarkHandler)
	statsHandlers := httpInfra.NewStatsHandlers(getHabitStatsHandler)
	healthHandlers := httpInfra.NewHealthHandlers(db.Conn())

	router := httpInfra.NewRouter(cfg.CORSOrigins, habitHandlers, authHandlers, statsHandlers, healthHandlers, jwtService)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
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
