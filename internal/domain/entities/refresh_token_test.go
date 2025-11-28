package entities

import (
	"testing"
	"time"
)

func TestNewRefreshToken(t *testing.T) {
	userID := "user-123"
	token := "refresh-token-abc"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	rt := NewRefreshToken(userID, token, expiresAt)

	if rt == nil {
		t.Fatal("NewRefreshToken() returned nil")
	}

	if rt.UserID != userID {
		t.Errorf("UserID = %v, want %v", rt.UserID, userID)
	}

	if rt.Token != token {
		t.Errorf("Token = %v, want %v", rt.Token, token)
	}

	if !rt.ExpiresAt.Equal(expiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", rt.ExpiresAt, expiresAt)
	}

	if rt.RevokedAt != nil {
		t.Error("RevokedAt should be nil for new token")
	}

	if rt.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		revoked   bool
		want      bool
	}{
		{
			name:      "valid token",
			expiresAt: time.Now().Add(24 * time.Hour),
			revoked:   false,
			want:      true,
		},
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revoked:   false,
			want:      false,
		},
		{
			name:      "revoked token",
			expiresAt: time.Now().Add(24 * time.Hour),
			revoked:   true,
			want:      false,
		},
		{
			name:      "expired and revoked",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revoked:   true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := NewRefreshToken("user-123", "token", tt.expiresAt)
			if tt.revoked {
				rt.Revoke()
			}

			got := rt.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefreshToken_Revoke(t *testing.T) {
	rt := NewRefreshToken("user-123", "token", time.Now().Add(24*time.Hour))

	if rt.RevokedAt != nil {
		t.Error("RevokedAt should be nil before Revoke()")
	}

	if !rt.IsValid() {
		t.Error("Token should be valid before Revoke()")
	}

	rt.Revoke()

	if rt.RevokedAt == nil {
		t.Fatal("RevokedAt should not be nil after Revoke()")
	}

	if rt.IsValid() {
		t.Error("Token should not be valid after Revoke()")
	}

	now := time.Now()
	diff := now.Sub(*rt.RevokedAt)
	if diff > time.Second || diff < 0 {
		t.Errorf("RevokedAt difference too large: %v", diff)
	}
}

func TestRefreshToken_RevokeMultipleTimes(t *testing.T) {
	rt := NewRefreshToken("user-123", "token", time.Now().Add(24*time.Hour))

	rt.Revoke()
	firstRevokedAt := rt.RevokedAt

	time.Sleep(10 * time.Millisecond)
	rt.Revoke()

	if rt.RevokedAt == firstRevokedAt {
		t.Error("Revoke() should update RevokedAt on subsequent calls")
	}
}

func TestRefreshToken_ExpirationCheck(t *testing.T) {
	expiresIn := 100 * time.Millisecond
	rt := NewRefreshToken("user-123", "token", time.Now().Add(expiresIn))

	if !rt.IsValid() {
		t.Error("Token should be valid initially")
	}

	time.Sleep(expiresIn + 10*time.Millisecond)

	if rt.IsValid() {
		t.Error("Token should be invalid after expiration")
	}
}

func TestRefreshToken_MultipleTokens(t *testing.T) {
	token1 := NewRefreshToken("user-1", "token-1", time.Now().Add(24*time.Hour))
	token2 := NewRefreshToken("user-2", "token-2", time.Now().Add(24*time.Hour))

	if token1.UserID == token2.UserID {
		t.Error("UserIDs should be different")
	}

	if token1.Token == token2.Token {
		t.Error("Tokens should be different")
	}

	if token1.CreatedAt.Equal(token2.CreatedAt) {
		t.Log("Warning: CreatedAt timestamps are identical (possible race condition in test)")
	}
}
