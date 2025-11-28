package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DBPath != "/test/db.sqlite" {
		t.Errorf("DBPath = %v, want %v", cfg.DBPath, "/test/db.sqlite")
	}
	if cfg.AppURL != "http://test.com" {
		t.Errorf("AppURL = %v, want %v", cfg.AppURL, "http://test.com")
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("JWTSecret = %v, want %v", cfg.JWTSecret, "test-secret")
	}
	if cfg.JWTExpiry != "1h" {
		t.Errorf("JWTExpiry = %v, want %v", cfg.JWTExpiry, "1h")
	}
	if cfg.RefreshTokenExpiry != "7d" {
		t.Errorf("RefreshTokenExpiry = %v, want %v", cfg.RefreshTokenExpiry, "7d")
	}
	if cfg.DefaultTimezone != "UTC" {
		t.Errorf("DefaultTimezone = %v, want %v", cfg.DefaultTimezone, "UTC")
	}
}

func TestLoad_WithDefaults(t *testing.T) {
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %v, want default %v", cfg.Port, "8080")
	}

	if cfg.SMTPPort != "587" {
		t.Errorf("SMTPPort = %v, want default %v", cfg.SMTPPort, "587")
	}

	if cfg.SupportEmail != "contact@apocapoc.app" {
		t.Errorf("SupportEmail = %v, want default %v", cfg.SupportEmail, "contact@apocapoc.app")
	}

	if cfg.SendWelcomeEmail != "false" {
		t.Errorf("SendWelcomeEmail = %v, want default %v", cfg.SendWelcomeEmail, "false")
	}

	if cfg.RegistrationMode != "open" {
		t.Errorf("RegistrationMode = %v, want default %v", cfg.RegistrationMode, "open")
	}
}

func TestLoad_WithCustomDefaults(t *testing.T) {
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	os.Setenv("PORT", "3000")
	os.Setenv("SMTP_PORT", "465")
	os.Setenv("SUPPORT_EMAIL", "support@test.com")
	os.Setenv("SEND_WELCOME_EMAIL", "true")
	os.Setenv("REGISTRATION_MODE", "closed")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "3000" {
		t.Errorf("Port = %v, want %v", cfg.Port, "3000")
	}

	if cfg.SMTPPort != "465" {
		t.Errorf("SMTPPort = %v, want %v", cfg.SMTPPort, "465")
	}

	if cfg.SupportEmail != "support@test.com" {
		t.Errorf("SupportEmail = %v, want %v", cfg.SupportEmail, "support@test.com")
	}

	if cfg.SendWelcomeEmail != "true" {
		t.Errorf("SendWelcomeEmail = %v, want %v", cfg.SendWelcomeEmail, "true")
	}

	if cfg.RegistrationMode != "closed" {
		t.Errorf("RegistrationMode = %v, want %v", cfg.RegistrationMode, "closed")
	}
}

func TestLoad_MissingDBPath(t *testing.T) {
	clearEnv()
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing DB_PATH but got nil")
	}

	expectedMsg := "DB_PATH is required"
	if err.Error() != expectedMsg {
		t.Errorf("Load() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv()
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing JWT_SECRET but got nil")
	}

	expectedMsg := "JWT_SECRET is required"
	if err.Error() != expectedMsg {
		t.Errorf("Load() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_MissingJWTExpiry(t *testing.T) {
	clearEnv()
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing JWT_EXPIRY but got nil")
	}

	expectedMsg := "JWT_EXPIRY is required"
	if err.Error() != expectedMsg {
		t.Errorf("Load() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_MissingRefreshTokenExpiry(t *testing.T) {
	clearEnv()
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing REFRESH_TOKEN_EXPIRY but got nil")
	}

	expectedMsg := "REFRESH_TOKEN_EXPIRY is required"
	if err.Error() != expectedMsg {
		t.Errorf("Load() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_MissingDefaultTimezone(t *testing.T) {
	clearEnv()
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing DEFAULT_TIMEZONE but got nil")
	}

	expectedMsg := "DEFAULT_TIMEZONE is required"
	if err.Error() != expectedMsg {
		t.Errorf("Load() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoad_WithSMTPConfig(t *testing.T) {
	os.Setenv("DB_PATH", "/test/db.sqlite")
	os.Setenv("APP_URL", "http://test.com")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "7d")
	os.Setenv("DEFAULT_TIMEZONE", "UTC")
	os.Setenv("SMTP_HOST", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "user@test.com")
	os.Setenv("SMTP_PASSWORD", "test-password")
	os.Setenv("SMTP_FROM", "noreply@test.com")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SMTPHost != "smtp.test.com" {
		t.Errorf("SMTPHost = %v, want %v", cfg.SMTPHost, "smtp.test.com")
	}
	if cfg.SMTPPort != "587" {
		t.Errorf("SMTPPort = %v, want %v", cfg.SMTPPort, "587")
	}
	if cfg.SMTPUser != "user@test.com" {
		t.Errorf("SMTPUser = %v, want %v", cfg.SMTPUser, "user@test.com")
	}
	if cfg.SMTPPassword != "test-password" {
		t.Errorf("SMTPPassword = %v, want %v", cfg.SMTPPassword, "test-password")
	}
	if cfg.SMTPFrom != "noreply@test.com" {
		t.Errorf("SMTPFrom = %v, want %v", cfg.SMTPFrom, "noreply@test.com")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "env value exists",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "env value empty uses default",
			key:          "TEST_KEY_EMPTY",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
		{
			name:         "env value not set uses default",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvOrDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func clearEnv() {
	os.Unsetenv("DB_PATH")
	os.Unsetenv("PORT")
	os.Unsetenv("APP_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRY")
	os.Unsetenv("REFRESH_TOKEN_EXPIRY")
	os.Unsetenv("DEFAULT_TIMEZONE")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("SMTP_FROM")
	os.Unsetenv("SUPPORT_EMAIL")
	os.Unsetenv("SEND_WELCOME_EMAIL")
	os.Unsetenv("REGISTRATION_MODE")
}
