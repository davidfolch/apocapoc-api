package entities

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	timezone := "Europe/Madrid"

	user := NewUser(email, passwordHash, timezone)

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	if user.PasswordHash != passwordHash {
		t.Errorf("Expected password hash %s, got %s", passwordHash, user.PasswordHash)
	}

	if user.Timezone != timezone {
		t.Errorf("Expected timezone %s, got %s", timezone, user.Timezone)
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	diff := user.UpdatedAt.Sub(user.CreatedAt)
	if diff < 0 || diff > time.Second {
		t.Errorf("CreatedAt and UpdatedAt should be nearly identical, diff: %v", diff)
	}
}

func TestUser_DefaultTimezone(t *testing.T) {
	user := NewUser("test@example.com", "hash", "")

	if user.Timezone != "UTC" {
		t.Errorf("Expected default timezone UTC, got %s", user.Timezone)
	}
}
