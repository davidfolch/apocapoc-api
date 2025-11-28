package crypto

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptHasher(t *testing.T) {
	hasher := NewBcryptHasher()
	if hasher == nil {
		t.Fatal("NewBcryptHasher() returned nil")
	}

	_, ok := hasher.(*BcryptHasher)
	if !ok {
		t.Fatal("NewBcryptHasher() did not return *BcryptHasher")
	}
}

func TestBcryptHasher_Hash(t *testing.T) {
	hasher := NewBcryptHasher()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!123$%^&*()",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "unicode password",
			password: "pässwörd123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if hashed == "" {
				t.Fatal("Hash() returned empty string")
			}

			if hashed == tt.password {
				t.Fatal("Hash() returned the same as input password")
			}

			err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(tt.password))
			if err != nil {
				t.Fatalf("Generated hash does not match original password: %v", err)
			}
		})
	}
}

func TestBcryptHasher_Compare(t *testing.T) {
	hasher := NewBcryptHasher()

	tests := []struct {
		name           string
		password       string
		compareWith    string
		expectError    bool
		errorAssertion func(error) bool
	}{
		{
			name:        "matching passwords",
			password:    "password123",
			compareWith: "password123",
			expectError: false,
		},
		{
			name:        "non-matching passwords",
			password:    "password123",
			compareWith: "wrongpassword",
			expectError: true,
			errorAssertion: func(err error) bool {
				return err == bcrypt.ErrMismatchedHashAndPassword
			},
		},
		{
			name:        "empty password comparison",
			password:    "",
			compareWith: "",
			expectError: false,
		},
		{
			name:        "unicode password match",
			password:    "pässwörd123",
			compareWith: "pässwörd123",
			expectError: false,
		},
		{
			name:        "unicode password mismatch",
			password:    "pässwörd123",
			compareWith: "password123",
			expectError: true,
			errorAssertion: func(err error) bool {
				return err == bcrypt.ErrMismatchedHashAndPassword
			},
		},
		{
			name:        "case sensitive",
			password:    "Password123",
			compareWith: "password123",
			expectError: true,
			errorAssertion: func(err error) bool {
				return err == bcrypt.ErrMismatchedHashAndPassword
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := hasher.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			err = hasher.Compare(hashed, tt.compareWith)
			if tt.expectError {
				if err == nil {
					t.Fatal("Compare() expected error but got nil")
				}
				if tt.errorAssertion != nil && !tt.errorAssertion(err) {
					t.Fatalf("Compare() error = %v, but assertion failed", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Compare() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestBcryptHasher_CompareWithInvalidHash(t *testing.T) {
	hasher := NewBcryptHasher()

	tests := []struct {
		name        string
		invalidHash string
		password    string
		expectError bool
	}{
		{
			name:        "invalid hash format",
			invalidHash: "not-a-valid-hash",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "empty hash",
			invalidHash: "",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "corrupted hash",
			invalidHash: "$2a$10$invalidhashdata",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.Compare(tt.invalidHash, tt.password)
			if !tt.expectError && err != nil {
				t.Fatalf("Compare() unexpected error = %v", err)
			}
			if tt.expectError && err == nil {
				t.Fatal("Compare() expected error but got nil")
			}
		})
	}
}

func TestBcryptHasher_HashGeneratesDifferentHashes(t *testing.T) {
	hasher := NewBcryptHasher()
	password := "samePassword123"

	hash1, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("Hash() generated identical hashes for same password (should use salt)")
	}

	if err := hasher.Compare(hash1, password); err != nil {
		t.Fatalf("First hash doesn't match password: %v", err)
	}

	if err := hasher.Compare(hash2, password); err != nil {
		t.Fatalf("Second hash doesn't match password: %v", err)
	}
}
