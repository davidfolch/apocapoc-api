package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		expiryHours int
	}{
		{
			name:        "standard configuration",
			secret:      "my-secret-key",
			expiryHours: 24,
		},
		{
			name:        "short expiry",
			secret:      "test-secret",
			expiryHours: 1,
		},
		{
			name:        "long expiry",
			secret:      "test-secret",
			expiryHours: 168,
		},
		{
			name:        "empty secret",
			secret:      "",
			expiryHours: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJWTService(tt.secret, tt.expiryHours)
			if service == nil {
				t.Fatal("NewJWTService() returned nil")
			}

			if string(service.secret) != tt.secret {
				t.Errorf("secret = %v, want %v", string(service.secret), tt.secret)
			}

			expectedExpiry := time.Duration(tt.expiryHours) * time.Hour
			if service.expiry != expectedExpiry {
				t.Errorf("expiry = %v, want %v", service.expiry, expectedExpiry)
			}
		})
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	service := NewJWTService("test-secret-key", 24)

	tests := []struct {
		name   string
		userID string
		email  string
	}{
		{
			name:   "standard user",
			userID: "user-123",
			email:  "user@example.com",
		},
		{
			name:   "empty user ID",
			userID: "",
			email:  "user@example.com",
		},
		{
			name:   "empty email",
			userID: "user-123",
			email:  "",
		},
		{
			name:   "both empty",
			userID: "",
			email:  "",
		},
		{
			name:   "special characters in email",
			userID: "user-456",
			email:  "user+test@example.com",
		},
		{
			name:   "uuid as user ID",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			email:  "uuid@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.email)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}

			if token == "" {
				t.Fatal("GenerateToken() returned empty token")
			}

			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() error = %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("UserID = %v, want %v", claims.UserID, tt.userID)
			}

			if claims.Email != tt.email {
				t.Errorf("Email = %v, want %v", claims.Email, tt.email)
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret, 24)

	tests := []struct {
		name        string
		setupToken  func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := service.GenerateToken("user-123", "user@example.com")
				return token
			},
			expectError: false,
		},
		{
			name: "invalid token format",
			setupToken: func() string {
				return "not.a.valid.token"
			},
			expectError: true,
		},
		{
			name: "empty token",
			setupToken: func() string {
				return ""
			},
			expectError: true,
		},
		{
			name: "token with wrong secret",
			setupToken: func() string {
				wrongService := NewJWTService("wrong-secret", 24)
				token, _ := wrongService.GenerateToken("user-123", "user@example.com")
				return token
			},
			expectError: true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				expiredService := NewJWTService(secret, -1)
				token, _ := expiredService.GenerateToken("user-123", "user@example.com")
				return token
			},
			expectError: true,
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed"
			},
			expectError: true,
		},
		{
			name: "token with invalid signature",
			setupToken: func() string {
				token, _ := service.GenerateToken("user-123", "user@example.com")
				return token[:len(token)-5] + "xxxxx"
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := service.ValidateToken(token)

			if tt.expectError {
				if err == nil {
					t.Fatal("ValidateToken() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("ValidateToken() unexpected error = %v", err)
				}
				if claims == nil {
					t.Fatal("ValidateToken() returned nil claims")
				}
			}
		})
	}
}

func TestJWTService_ValidateTokenClaims(t *testing.T) {
	service := NewJWTService("test-secret-key", 24)

	userID := "user-123"
	email := "user@example.com"

	token, err := service.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("Email = %v, want %v", claims.Email, email)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}

	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt is nil")
	}

	if claims.ExpiresAt.Before(claims.IssuedAt.Time) {
		t.Error("ExpiresAt is before IssuedAt")
	}

	expectedExpiry := claims.IssuedAt.Add(24 * time.Hour)
	if !claims.ExpiresAt.Time.Equal(expectedExpiry) {
		diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
		if diff > time.Second || diff < -time.Second {
			t.Errorf("ExpiresAt = %v, want approximately %v (diff: %v)", claims.ExpiresAt.Time, expectedExpiry, diff)
		}
	}
}

func TestJWTService_TokenExpiry(t *testing.T) {
	tests := []struct {
		name        string
		expiryHours int
	}{
		{
			name:        "1 hour expiry",
			expiryHours: 1,
		},
		{
			name:        "24 hours expiry",
			expiryHours: 24,
		},
		{
			name:        "168 hours (1 week) expiry",
			expiryHours: 168,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJWTService("test-secret", tt.expiryHours)
			token, err := service.GenerateToken("user-123", "user@example.com")
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}

			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() error = %v", err)
			}

			expectedExpiry := time.Now().Add(time.Duration(tt.expiryHours) * time.Hour)
			diff := claims.ExpiresAt.Time.Sub(expectedExpiry)

			if diff > time.Second || diff < -time.Second {
				t.Errorf("ExpiresAt difference too large: %v", diff)
			}
		})
	}
}

func TestJWTService_ValidateTokenWithWrongSigningMethod(t *testing.T) {
	service := NewJWTService("test-secret", 24)

	claims := &Claims{
		UserID: "user-123",
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = service.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("ValidateToken() expected error for none signing method but got nil")
	}
}

func TestJWTService_MultipleTokens(t *testing.T) {
	service := NewJWTService("test-secret", 24)

	token1, err := service.GenerateToken("user-1", "user1@example.com")
	if err != nil {
		t.Fatalf("GenerateToken(1) error = %v", err)
	}

	token2, err := service.GenerateToken("user-2", "user2@example.com")
	if err != nil {
		t.Fatalf("GenerateToken(2) error = %v", err)
	}

	if token1 == token2 {
		t.Error("Generated identical tokens for different users")
	}

	claims1, err := service.ValidateToken(token1)
	if err != nil {
		t.Fatalf("ValidateToken(1) error = %v", err)
	}

	claims2, err := service.ValidateToken(token2)
	if err != nil {
		t.Fatalf("ValidateToken(2) error = %v", err)
	}

	if claims1.UserID == claims2.UserID {
		t.Error("Claims have same UserID")
	}

	if claims1.Email == claims2.Email {
		t.Error("Claims have same Email")
	}
}
