package entities

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func NewRefreshToken(userID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		RevokedAt: nil,
	}
}

func (rt *RefreshToken) IsValid() bool {
	if rt.RevokedAt != nil {
		return false
	}
	return time.Now().Before(rt.ExpiresAt)
}

func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
}
