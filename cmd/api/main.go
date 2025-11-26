package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"habit-tracker-api/internal/application/commands"
	"habit-tracker-api/internal/application/queries"
	"habit-tracker-api/internal/infrastructure/auth"
	"habit-tracker-api/internal/infrastructure/config"
	httpInfra "habit-tracker-api/internal/infrastructure/http"
	"habit-tracker-api/internal/infrastructure/persistence/sqlite"
)

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

	jwtService := auth.NewJWTService(cfg.JWTSecret, jwtExpiryHours)

	userRepo := sqlite.NewUserRepository(db.Conn())
	habitRepo := sqlite.NewHabitRepository(db.Conn())
	entryRepo := sqlite.NewHabitEntryRepository(db.Conn())

	registerHandler := commands.NewRegisterUserHandler(userRepo)
	loginHandler := queries.NewLoginUserHandler(userRepo)
	createHandler := commands.NewCreateHabitHandler(habitRepo)
	getTodaysHandler := queries.NewGetTodaysHabitsHandler(habitRepo, entryRepo)
	markHandler := commands.NewMarkHabitHandler(entryRepo, habitRepo)

	authHandlers := httpInfra.NewAuthHandlers(registerHandler, loginHandler, jwtService)
	habitHandlers := httpInfra.NewHabitHandlers(createHandler, getTodaysHandler, markHandler)

	router := httpInfra.NewRouter(cfg.CORSOrigins, habitHandlers, authHandlers, jwtService)

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
