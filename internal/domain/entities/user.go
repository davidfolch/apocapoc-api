package entities

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Timezone     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, passwordHash, timezone string) *User {
	now := time.Now()
	if timezone == "" {
		timezone = "UTC"
	}
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		Timezone:     timezone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
