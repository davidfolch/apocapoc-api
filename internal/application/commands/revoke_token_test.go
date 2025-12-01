package commands

import (
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockRefreshTokenRepo struct {
	tokens map[string]*entities.RefreshToken
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *entities.RefreshToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *mockRefreshTokenRepo) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockRefreshTokenRepo) FindByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	var tokens []*entities.RefreshToken
	for _, t := range m.tokens {
		if t.UserID == userID {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (m *mockRefreshTokenRepo) RevokeByToken(ctx context.Context, token string) error {
	if t, ok := m.tokens[token]; ok {
		t.Revoke()
		return nil
	}
	return errors.ErrNotFound
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoke()
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

func TestRevokeTokenHandler_Handle(t *testing.T) {
	repo := &mockRefreshTokenRepo{
		tokens: make(map[string]*entities.RefreshToken),
	}

	handler := NewRevokeTokenHandler(repo)

	token := entities.NewRefreshToken("user-123", "valid-token", time.Now().Add(24*time.Hour))
	repo.Create(context.Background(), token)

	tests := []struct {
		name        string
		cmd         RevokeTokenCommand
		expectError bool
		expectedErr error
	}{
		{
			name: "revoke valid token",
			cmd: RevokeTokenCommand{
				RefreshToken: "valid-token",
			},
			expectError: false,
		},
		{
			name: "revoke empty token",
			cmd: RevokeTokenCommand{
				RefreshToken: "",
			},
			expectError: true,
			expectedErr: errors.ErrInvalidInput,
		},
		{
			name: "revoke non-existent token",
			cmd: RevokeTokenCommand{
				RefreshToken: "non-existent-token",
			},
			expectError: true,
			expectedErr: errors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.Background(), tt.cmd)

			if tt.expectError {
				if err == nil {
					t.Fatal("Handle() expected error but got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Handle() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Fatalf("Handle() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestRevokeAllTokensHandler_Handle(t *testing.T) {
	repo := &mockRefreshTokenRepo{
		tokens: make(map[string]*entities.RefreshToken),
	}

	handler := NewRevokeAllTokensHandler(repo)

	token1 := entities.NewRefreshToken("user-123", "token-1", time.Now().Add(24*time.Hour))
	token2 := entities.NewRefreshToken("user-123", "token-2", time.Now().Add(24*time.Hour))
	token3 := entities.NewRefreshToken("user-456", "token-3", time.Now().Add(24*time.Hour))

	repo.Create(context.Background(), token1)
	repo.Create(context.Background(), token2)
	repo.Create(context.Background(), token3)

	tests := []struct {
		name        string
		cmd         RevokeAllTokensCommand
		expectError bool
		expectedErr error
	}{
		{
			name: "revoke all tokens for user",
			cmd: RevokeAllTokensCommand{
				UserID: "user-123",
			},
			expectError: false,
		},
		{
			name: "revoke with empty user ID",
			cmd: RevokeAllTokensCommand{
				UserID: "",
			},
			expectError: true,
			expectedErr: errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.Background(), tt.cmd)

			if tt.expectError {
				if err == nil {
					t.Fatal("Handle() expected error but got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Handle() error = %v, want %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Fatalf("Handle() unexpected error = %v", err)
				}

				if !tt.expectError && tt.cmd.UserID == "user-123" {
					if repo.tokens["token-1"].RevokedAt == nil {
						t.Error("token-1 should be revoked")
					}
					if repo.tokens["token-2"].RevokedAt == nil {
						t.Error("token-2 should be revoked")
					}
					if repo.tokens["token-3"].RevokedAt != nil {
						t.Error("token-3 should not be revoked")
					}
				}
			}
		})
	}
}

func (m *mockRefreshTokenRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
