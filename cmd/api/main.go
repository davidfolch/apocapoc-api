package main

import (
	"fmt"
	"log"
	"net/http"

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

	router := httpInfra.NewRouter(cfg.CORSOrigins)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
