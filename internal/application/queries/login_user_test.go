package queries

import (
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockLoginUserRepo struct {
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
}

func (m *mockLoginUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, errors.ErrNotFound
}

func (m *mockLoginUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockLoginUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockLoginUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockLoginUserRepo) Update(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockLoginUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockLoginPasswordHasher struct {
	compareFunc func(hashedPassword, password string) error
}

func (m *mockLoginPasswordHasher) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *mockLoginPasswordHasher) Compare(hashedPassword, password string) error {
	if m.compareFunc != nil {
		return m.compareFunc(hashedPassword, password)
	}
	return nil
}

func TestLoginUserHandler_Success(t *testing.T) {
	user := entities.NewUser("test@example.com", "hashed_password", "UTC")
	user.ID = "user-123"
	user.EmailVerified = true

	repo := &mockLoginUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return user, nil
		},
	}

	hasher := &mockLoginPasswordHasher{
		compareFunc: func(hashedPassword, password string) error {
			return nil
		},
	}

	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := handler.Handle(context.Background(), query)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if result.UserID != "user-123" {
		t.Errorf("UserID = %v, want %v", result.UserID, "user-123")
	}

	if result.Email != "test@example.com" {
		t.Errorf("Email = %v, want %v", result.Email, "test@example.com")
	}

	if result.Timezone != "UTC" {
		t.Errorf("Timezone = %v, want %v", result.Timezone, "UTC")
	}
}

func TestLoginUserHandler_EmptyEmail(t *testing.T) {
	repo := &mockLoginUserRepo{}
	hasher := &mockLoginPasswordHasher{}
	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "",
		Password: "password123",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestLoginUserHandler_EmptyPassword(t *testing.T) {
	repo := &mockLoginUserRepo{}
	hasher := &mockLoginPasswordHasher{}
	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "test@example.com",
		Password: "",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestLoginUserHandler_UserNotFound(t *testing.T) {
	repo := &mockLoginUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, errors.ErrNotFound
		},
	}

	hasher := &mockLoginPasswordHasher{}
	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrNotFound {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrNotFound)
	}
}

func TestLoginUserHandler_InvalidPassword(t *testing.T) {
	user := entities.NewUser("test@example.com", "hashed_password", "UTC")
	user.ID = "user-123"
	user.EmailVerified = true

	repo := &mockLoginUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return user, nil
		},
	}

	hasher := &mockLoginPasswordHasher{
		compareFunc: func(hashedPassword, password string) error {
			return errors.ErrInvalidInput
		},
	}

	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrNotFound {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrNotFound)
	}
}

func TestLoginUserHandler_EmailNotVerified(t *testing.T) {
	user := entities.NewUser("test@example.com", "hashed_password", "UTC")
	user.ID = "user-123"
	user.EmailVerified = false

	repo := &mockLoginUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return user, nil
		},
	}

	hasher := &mockLoginPasswordHasher{
		compareFunc: func(hashedPassword, password string) error {
			return nil
		},
	}

	handler := NewLoginUserHandler(repo, hasher)

	query := LoginUserQuery{
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrEmailNotVerified {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrEmailNotVerified)
	}
}
