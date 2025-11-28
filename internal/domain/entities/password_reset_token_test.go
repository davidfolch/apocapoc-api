package entities

import (
	"testing"
	"time"
)

func TestNewPasswordResetToken(t *testing.T) {
	userID := "user-123"
	token := "reset-token-abc"
	expiresAt := time.Now().Add(1 * time.Hour)

	prt := NewPasswordResetToken(userID, token, expiresAt)

	if prt == nil {
		t.Fatal("NewPasswordResetToken() returned nil")
	}

	if prt.ID == "" {
		t.Error("ID is empty")
	}

	if prt.UserID != userID {
		t.Errorf("UserID = %v, want %v", prt.UserID, userID)
	}

	if prt.Token != token {
		t.Errorf("Token = %v, want %v", prt.Token, token)
	}

	if !prt.ExpiresAt.Equal(expiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", prt.ExpiresAt, expiresAt)
	}

	if prt.UsedAt != nil {
		t.Error("UsedAt should be nil for new token")
	}

	if prt.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "future expiry",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "past expiry",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expires in 1 second",
			expiresAt: time.Now().Add(1 * time.Second),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prt := NewPasswordResetToken("user-123", "token", tt.expiresAt)
			got := prt.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	prt := NewPasswordResetToken("user-123", "token", time.Now().Add(1*time.Hour))

	if prt.IsUsed() {
		t.Error("IsUsed() should return false for new token")
	}

	prt.MarkAsUsed()

	if !prt.IsUsed() {
		t.Error("IsUsed() should return true after MarkAsUsed()")
	}
}

func TestPasswordResetToken_MarkAsUsed(t *testing.T) {
	prt := NewPasswordResetToken("user-123", "token", time.Now().Add(1*time.Hour))

	if prt.UsedAt != nil {
		t.Error("UsedAt should be nil before MarkAsUsed()")
	}

	prt.MarkAsUsed()

	if prt.UsedAt == nil {
		t.Fatal("UsedAt should not be nil after MarkAsUsed()")
	}

	now := time.Now()
	diff := now.Sub(*prt.UsedAt)
	if diff > time.Second || diff < 0 {
		t.Errorf("UsedAt difference too large: %v", diff)
	}

	firstUsedAt := prt.UsedAt
	time.Sleep(10 * time.Millisecond)
	prt.MarkAsUsed()

	if prt.UsedAt == firstUsedAt {
		t.Error("MarkAsUsed() should update UsedAt on subsequent calls")
	}
}

func TestPasswordResetToken_MultipleTokens(t *testing.T) {
	token1 := NewPasswordResetToken("user-1", "token-1", time.Now().Add(1*time.Hour))
	token2 := NewPasswordResetToken("user-2", "token-2", time.Now().Add(1*time.Hour))

	if token1.ID == token2.ID {
		t.Error("NewPasswordResetToken() generated identical IDs")
	}

	if token1.UserID == token2.UserID {
		t.Error("UserIDs should be different")
	}

	if token1.Token == token2.Token {
		t.Error("Tokens should be different")
	}
}
