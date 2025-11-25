package main

import (
	"fmt"
	"log"
	"net/http"

	"habit-tracker-api/internal/application/commands"
	"habit-tracker-api/internal/application/queries"
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

	habitRepo := sqlite.NewHabitRepository(db.Conn())
	entryRepo := sqlite.NewHabitEntryRepository(db.Conn())

	createHandler := commands.NewCreateHabitHandler(habitRepo)
	getTodaysHandler := queries.NewGetTodaysHabitsHandler(habitRepo, entryRepo)
	markHandler := commands.NewMarkHabitHandler(entryRepo, habitRepo)

	habitHandlers := httpInfra.NewHabitHandlers(createHandler, getTodaysHandler, markHandler)

	router := httpInfra.NewRouter(cfg.CORSOrigins, habitHandlers)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
