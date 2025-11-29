package entities

import "time"

type User struct {
	ID                      string
	Email                   string
	PasswordHash            string
	EmailVerified           bool
	EmailVerificationToken  *string
	EmailVerificationExpiry *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func NewUser(email, passwordHash string) *User {
	now := time.Now()
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
