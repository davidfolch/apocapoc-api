package commands

import (
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockResetPasswordUserRepo struct {
	findByIDFunc func(ctx context.Context, id string) (*entities.User, error)
	updateFunc   func(ctx context.Context, user *entities.User) error
	users        map[string]*entities.User
}

func (m *mockResetPasswordUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockResetPasswordUserRepo) Update(ctx context.Context, user *entities.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockResetPasswordUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockResetPasswordUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockResetPasswordUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockResetPasswordUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockPasswordResetTokenRepo struct {
	findByTokenFunc func(ctx context.Context, token string) (*entities.PasswordResetToken, error)
	updateFunc      func(ctx context.Context, token *entities.PasswordResetToken) error
	tokens          map[string]*entities.PasswordResetToken
}

func (m *mockPasswordResetTokenRepo) Create(ctx context.Context, token *entities.PasswordResetToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *mockPasswordResetTokenRepo) FindByToken(ctx context.Context, token string) (*entities.PasswordResetToken, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	if t, ok := m.tokens[token]; ok {
		return t, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockPasswordResetTokenRepo) Update(ctx context.Context, token *entities.PasswordResetToken) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, token)
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockPasswordResetTokenRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockResetPasswordHasher struct {
	hashFunc func(password string) (string, error)
}

func (m *mockResetPasswordHasher) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *mockResetPasswordHasher) Compare(hashedPassword, password string) error {
	return nil
}

func TestResetPasswordHandler_Success(t *testing.T) {
	user := entities.NewUser("test@example.com", "old_hash", "UTC")
	user.ID = "user-123"

	resetToken := entities.NewPasswordResetToken(
		user.ID,
		"reset-token",
		time.Now().Add(1*time.Hour),
	)

	var updatedUser *entities.User
	var updatedToken *entities.PasswordResetToken

	userRepo := &mockResetPasswordUserRepo{
		users: map[string]*entities.User{
			user.ID: user,
		},
		updateFunc: func(ctx context.Context, u *entities.User) error {
			updatedUser = u
			return nil
		},
	}

	tokenRepo := &mockPasswordResetTokenRepo{
		tokens: map[string]*entities.PasswordResetToken{
			resetToken.Token: resetToken,
		},
		updateFunc: func(ctx context.Context, t *entities.PasswordResetToken) error {
			updatedToken = t
			return nil
		},
	}

	hasher := &mockResetPasswordHasher{}

	handler := NewResetPasswordHandler(userRepo, tokenRepo, hasher)

	cmd := ResetPasswordCommand{
		Token:       "reset-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if updatedUser == nil {
		t.Fatal("User was not updated")
	}

	if updatedUser.PasswordHash != "hashed_NewP@ssw0rd123" {
		t.Errorf("PasswordHash = %v, want %v", updatedUser.PasswordHash, "hashed_NewP@ssw0rd123")
	}

	if updatedToken == nil {
		t.Fatal("Token was not updated")
	}

	if !updatedToken.IsUsed() {
		t.Error("Token should be marked as used")
	}
}

func TestResetPasswordHandler_EmptyToken(t *testing.T) {
	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		&mockPasswordResetTokenRepo{tokens: make(map[string]*entities.PasswordResetToken)},
		&mockResetPasswordHasher{},
	)

	cmd := ResetPasswordCommand{
		Token:       "",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestResetPasswordHandler_EmptyPassword(t *testing.T) {
	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		&mockPasswordResetTokenRepo{tokens: make(map[string]*entities.PasswordResetToken)},
		&mockResetPasswordHasher{},
	)

	cmd := ResetPasswordCommand{
		Token:       "reset-token",
		NewPassword: "",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestResetPasswordHandler_InvalidPassword(t *testing.T) {
	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		&mockPasswordResetTokenRepo{tokens: make(map[string]*entities.PasswordResetToken)},
		&mockResetPasswordHasher{},
	)

	tests := []struct {
		name     string
		password string
	}{
		{"too short", "Short1!"},
		{"no uppercase", "password123!"},
		{"no lowercase", "PASSWORD123!"},
		{"no digit", "Password!"},
		{"no special char", "Password123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ResetPasswordCommand{
				Token:       "reset-token",
				NewPassword: tt.password,
			}

			err := handler.Handle(context.Background(), cmd)
			if err != errors.ErrInvalidInput {
				t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
			}
		})
	}
}

func TestResetPasswordHandler_TokenNotFound(t *testing.T) {
	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		&mockPasswordResetTokenRepo{tokens: make(map[string]*entities.PasswordResetToken)},
		&mockResetPasswordHasher{},
	)

	cmd := ResetPasswordCommand{
		Token:       "non-existent-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestResetPasswordHandler_ExpiredToken(t *testing.T) {
	resetToken := entities.NewPasswordResetToken(
		"user-123",
		"expired-token",
		time.Now().Add(-1*time.Hour),
	)

	tokenRepo := &mockPasswordResetTokenRepo{
		tokens: map[string]*entities.PasswordResetToken{
			resetToken.Token: resetToken,
		},
	}

	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		tokenRepo,
		&mockResetPasswordHasher{},
	)

	cmd := ResetPasswordCommand{
		Token:       "expired-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestResetPasswordHandler_UsedToken(t *testing.T) {
	resetToken := entities.NewPasswordResetToken(
		"user-123",
		"used-token",
		time.Now().Add(1*time.Hour),
	)
	resetToken.MarkAsUsed()

	tokenRepo := &mockPasswordResetTokenRepo{
		tokens: map[string]*entities.PasswordResetToken{
			resetToken.Token: resetToken,
		},
	}

	handler := NewResetPasswordHandler(
		&mockResetPasswordUserRepo{users: make(map[string]*entities.User)},
		tokenRepo,
		&mockResetPasswordHasher{},
	)

	cmd := ResetPasswordCommand{
		Token:       "used-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestResetPasswordHandler_UserNotFound(t *testing.T) {
	resetToken := entities.NewPasswordResetToken(
		"non-existent-user",
		"reset-token",
		time.Now().Add(1*time.Hour),
	)

	tokenRepo := &mockPasswordResetTokenRepo{
		tokens: map[string]*entities.PasswordResetToken{
			resetToken.Token: resetToken,
		},
	}

	userRepo := &mockResetPasswordUserRepo{
		users: make(map[string]*entities.User),
	}

	handler := NewResetPasswordHandler(userRepo, tokenRepo, &mockResetPasswordHasher{})

	cmd := ResetPasswordCommand{
		Token:       "reset-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrNotFound {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrNotFound)
	}
}

func TestResetPasswordHandler_HashingError(t *testing.T) {
	user := entities.NewUser("test@example.com", "old_hash", "UTC")
	user.ID = "user-123"

	resetToken := entities.NewPasswordResetToken(
		user.ID,
		"reset-token",
		time.Now().Add(1*time.Hour),
	)

	userRepo := &mockResetPasswordUserRepo{
		users: map[string]*entities.User{
			user.ID: user,
		},
	}

	tokenRepo := &mockPasswordResetTokenRepo{
		tokens: map[string]*entities.PasswordResetToken{
			resetToken.Token: resetToken,
		},
	}

	hasher := &mockResetPasswordHasher{
		hashFunc: func(password string) (string, error) {
			return "", errors.ErrInvalidInput
		},
	}

	handler := NewResetPasswordHandler(userRepo, tokenRepo, hasher)

	cmd := ResetPasswordCommand{
		Token:       "reset-token",
		NewPassword: "NewP@ssw0rd123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("Handle() expected error but got nil")
	}
}
