package entities

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func NewPasswordResetToken(userID, token string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
}

func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
}
