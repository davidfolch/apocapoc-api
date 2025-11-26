package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBType             string
	DBPath             string
	Port               string
	Host               string
	JWTSecret          string
	JWTExpiry          string
	RefreshTokenExpiry string
	CORSOrigins        string
	DefaultTimezone    string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		DBType:             os.Getenv("DB_TYPE"),
		DBPath:             os.Getenv("DB_PATH"),
		Port:               getEnvOrDefault("PORT", "8080"),
		Host:               getEnvOrDefault("HOST", "0.0.0.0"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTExpiry:          os.Getenv("JWT_EXPIRY"),
		RefreshTokenExpiry: os.Getenv("REFRESH_TOKEN_EXPIRY"),
		CORSOrigins:        os.Getenv("CORS_ORIGINS"),
		DefaultTimezone:    os.Getenv("DEFAULT_TIMEZONE"),
	}

	if cfg.DBType == "" {
		return nil, fmt.Errorf("DB_TYPE is required")
	}
	if cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.JWTExpiry == "" {
		return nil, fmt.Errorf("JWT_EXPIRY is required")
	}
	if cfg.RefreshTokenExpiry == "" {
		return nil, fmt.Errorf("REFRESH_TOKEN_EXPIRY is required")
	}
	if cfg.CORSOrigins == "" {
		return nil, fmt.Errorf("CORS_ORIGINS is required")
	}
	if cfg.DefaultTimezone == "" {
		return nil, fmt.Errorf("DEFAULT_TIMEZONE is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
