package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath              string
	Port                string
	AppURL              string
	JWTSecret           string
	JWTExpiry           string
	RefreshTokenExpiry  string
	DefaultTimezone     string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPassword        string
	SMTPFrom            string
	SupportEmail        string
	SendWelcomeEmail    string
	RegistrationMode    string
	LogLevel            string
	Environment         string
	BackupEnabled       string
	BackupInterval      string
	BackupRetentionDays string
	BackupPath          string
	BackupCompress      string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		DBPath:              os.Getenv("DB_PATH"),
		Port:                getEnvOrDefault("PORT", "8080"),
		AppURL:              getEnvOrDefault("APP_URL", "http://localhost:8080"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTExpiry:           os.Getenv("JWT_EXPIRY"),
		RefreshTokenExpiry:  os.Getenv("REFRESH_TOKEN_EXPIRY"),
		DefaultTimezone:     os.Getenv("DEFAULT_TIMEZONE"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:            os.Getenv("SMTP_FROM"),
		SupportEmail:        getEnvOrDefault("SUPPORT_EMAIL", "contact@apocapoc.app"),
		SendWelcomeEmail:    getEnvOrDefault("SEND_WELCOME_EMAIL", "false"),
		RegistrationMode:    getEnvOrDefault("REGISTRATION_MODE", "open"),
		LogLevel:            getEnvOrDefault("LOG_LEVEL", "info"),
		Environment:         getEnvOrDefault("ENVIRONMENT", "production"),
		BackupEnabled:       getEnvOrDefault("BACKUP_ENABLED", "false"),
		BackupInterval:      getEnvOrDefault("BACKUP_INTERVAL", "24h"),
		BackupRetentionDays: getEnvOrDefault("BACKUP_RETENTION_DAYS", "7"),
		BackupPath:          getEnvOrDefault("BACKUP_PATH", "./data/backups"),
		BackupCompress:      getEnvOrDefault("BACKUP_COMPRESS", "true"),
	}

	if cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH is required")
	}
	if cfg.AppURL == "" {
		return nil, fmt.Errorf("APP_URL is required")
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
