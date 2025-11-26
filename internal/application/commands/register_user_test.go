package commands

import (
	"context"
	"errors"
	"testing"

	"apocapoc-api/internal/domain/entities"
	appErrors "apocapoc-api/internal/shared/errors"
)

type mockUserRepo struct {
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
	createFunc      func(ctx context.Context, user *entities.User) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *entities.User) error {
	return nil
}

type mockPasswordHasher struct {
	hashFunc func(password string) (string, error)
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	return nil
}

func TestRegisterUserHandler_Success(t *testing.T) {
	var createdUser *entities.User
	repo := &mockUserRepo{
		createFunc: func(ctx context.Context, user *entities.User) error {
			user.ID = "test-user-id-123"
			createdUser = user
			return nil
		},
	}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	cmd := RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Secure123!",
		Timezone: "UTC",
	}

	userID, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userID == "" {
		t.Error("expected user ID, got empty string")
	}

	if createdUser == nil {
		t.Fatal("expected user to be created")
	}

	if createdUser.Email != cmd.Email {
		t.Errorf("expected email %q, got %q", cmd.Email, createdUser.Email)
	}

	if createdUser.Timezone != cmd.Timezone {
		t.Errorf("expected timezone %q, got %q", cmd.Timezone, createdUser.Timezone)
	}
}

func TestRegisterUserHandler_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"invalid format", "not-an-email"},
		{"missing @", "testexample.com"},
		{"missing domain", "test@"},
		{"spaces in email", "test user@example.com"},
		{"too long local part", string(make([]byte, 65)) + "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RegisterUserCommand{
				Email:    tt.email,
				Password: "Secure123!",
				Timezone: "UTC",
			}

			_, err := handler.Handle(context.Background(), cmd)
			if err != appErrors.ErrInvalidInput {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterUserHandler_InvalidPassword(t *testing.T) {
	repo := &mockUserRepo{}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	tests := []struct {
		name     string
		password string
	}{
		{"empty password", ""},
		{"too short", "Short1!"},
		{"no uppercase", "secure123!"},
		{"no lowercase", "SECURE123!"},
		{"no digit", "SecurePass!"},
		{"no special char", "SecurePass1"},
		{"only letters", "OnlyLetters"},
		{"only numbers", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RegisterUserCommand{
				Email:    "test@example.com",
				Password: tt.password,
				Timezone: "UTC",
			}

			_, err := handler.Handle(context.Background(), cmd)
			if err != appErrors.ErrInvalidInput {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterUserHandler_InvalidTimezone(t *testing.T) {
	repo := &mockUserRepo{}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	tests := []struct {
		name     string
		timezone string
	}{
		{"empty timezone", ""},
		{"invalid timezone", "InvalidTimezone"},
		{"numeric format", "GMT+1"},
		{"partial timezone", "America"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RegisterUserCommand{
				Email:    "test@example.com",
				Password: "Secure123!",
				Timezone: tt.timezone,
			}

			_, err := handler.Handle(context.Background(), cmd)
			if err != appErrors.ErrInvalidInput {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRegisterUserHandler_EmailAlreadyExists(t *testing.T) {
	existingUser := entities.NewUser("test@example.com", "hashed", "UTC")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return existingUser, nil
		},
	}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	cmd := RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Secure123!",
		Timezone: "UTC",
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err != appErrors.ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestRegisterUserHandler_PasswordHashingError(t *testing.T) {
	expectedErr := errors.New("hashing error")
	repo := &mockUserRepo{}
	hasher := &mockPasswordHasher{
		hashFunc: func(password string) (string, error) {
			return "", expectedErr
		},
	}
	handler := NewRegisterUserHandler(repo, hasher)

	cmd := RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Secure123!",
		Timezone: "UTC",
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err != expectedErr {
		t.Errorf("expected hashing error, got %v", err)
	}
}

func TestRegisterUserHandler_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")
	repo := &mockUserRepo{
		createFunc: func(ctx context.Context, user *entities.User) error {
			return expectedErr
		},
	}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	cmd := RegisterUserCommand{
		Email:    "test@example.com",
		Password: "Secure123!",
		Timezone: "UTC",
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err != expectedErr {
		t.Errorf("expected repository error, got %v", err)
	}
}

func TestRegisterUserHandler_EdgeCases(t *testing.T) {
	repo := &mockUserRepo{}
	hasher := &mockPasswordHasher{}
	handler := NewRegisterUserHandler(repo, hasher)

	tests := []struct {
		name     string
		cmd      RegisterUserCommand
		wantErr  error
	}{
		{
			"email with plus addressing",
			RegisterUserCommand{
				Email:    "user+tag@example.com",
				Password: "Secure123!",
				Timezone: "UTC",
			},
			nil,
		},
		{
			"email with subdomain",
			RegisterUserCommand{
				Email:    "user@mail.example.com",
				Password: "Secure123!",
				Timezone: "UTC",
			},
			nil,
		},
		{
			"complex timezone",
			RegisterUserCommand{
				Email:    "user@example.com",
				Password: "Secure123!",
				Timezone: "America/Argentina/Buenos_Aires",
			},
			nil,
		},
		{
			"password with unicode",
			RegisterUserCommand{
				Email:    "user@example.com",
				Password: "Sëcure123!",
				Timezone: "UTC",
			},
			nil,
		},
		{
			"very long valid password",
			RegisterUserCommand{
				Email:    "user@example.com",
				Password: "ValidP@ss1" + string(make([]byte, 100)),
				Timezone: "UTC",
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.cmd)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
