package commands

import (
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type mockVerifyEmailUserRepo struct {
	findByVerificationTokenFunc func(ctx context.Context, token string) (*entities.User, error)
	updateFunc                  func(ctx context.Context, user *entities.User) error
	users                       map[string]*entities.User
}

func (m *mockVerifyEmailUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	if m.findByVerificationTokenFunc != nil {
		return m.findByVerificationTokenFunc(ctx, token)
	}
	if user, ok := m.users[token]; ok {
		return user, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockVerifyEmailUserRepo) Update(ctx context.Context, user *entities.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockVerifyEmailUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockVerifyEmailUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockVerifyEmailUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockVerifyEmailUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockEmailService struct {
	sendFunc     func(message services.EmailMessage) error
	sentMessages []services.EmailMessage
}

func (m *mockEmailService) Send(message services.EmailMessage) error {
	m.sentMessages = append(m.sentMessages, message)
	if m.sendFunc != nil {
		return m.sendFunc(message)
	}
	return nil
}

func TestVerifyEmailHandler_Success(t *testing.T) {
	token := "valid-token"
	expiry := time.Now().Add(24 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	var updatedUser *entities.User
	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
		updateFunc: func(ctx context.Context, u *entities.User) error {
			updatedUser = u
			return nil
		},
	}

	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if updatedUser == nil {
		t.Fatal("User was not updated")
	}

	if !updatedUser.EmailVerified {
		t.Error("EmailVerified should be true")
	}

	if updatedUser.EmailVerificationToken != nil {
		t.Error("EmailVerificationToken should be nil after verification")
	}

	if updatedUser.EmailVerificationExpiry != nil {
		t.Error("EmailVerificationExpiry should be nil after verification")
	}
}

func TestVerifyEmailHandler_EmptyToken(t *testing.T) {
	repo := &mockVerifyEmailUserRepo{
		users: make(map[string]*entities.User),
	}
	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: "",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestVerifyEmailHandler_TokenNotFound(t *testing.T) {
	repo := &mockVerifyEmailUserRepo{
		users: make(map[string]*entities.User),
	}
	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: "non-existent-token",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestVerifyEmailHandler_AlreadyVerified(t *testing.T) {
	token := "valid-token"
	expiry := time.Now().Add(24 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = true
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
	}

	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrAlreadyExists {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrAlreadyExists)
	}
}

func TestVerifyEmailHandler_ExpiredToken(t *testing.T) {
	token := "expired-token"
	expiry := time.Now().Add(-1 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
	}

	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestVerifyEmailHandler_NilExpiry(t *testing.T) {
	token := "valid-token"

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = nil

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
	}

	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestVerifyEmailHandler_WithWelcomeEmail(t *testing.T) {
	token := "valid-token"
	expiry := time.Now().Add(24 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
	}

	emailService := &mockEmailService{}
	handler := NewVerifyEmailHandler(repo, emailService, true)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if len(emailService.sentMessages) != 1 {
		t.Fatalf("Expected 1 email sent, got %d", len(emailService.sentMessages))
	}

	sentEmail := emailService.sentMessages[0]
	if sentEmail.To != "test@example.com" {
		t.Errorf("Email To = %v, want %v", sentEmail.To, "test@example.com")
	}

	if sentEmail.Subject != "Welcome to Apocapoc!" {
		t.Errorf("Email Subject = %v, want %v", sentEmail.Subject, "Welcome to Apocapoc!")
	}

	if !sentEmail.IsHTML {
		t.Error("Email should be HTML")
	}
}

func TestVerifyEmailHandler_WithoutWelcomeEmail(t *testing.T) {
	token := "valid-token"
	expiry := time.Now().Add(24 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
	}

	emailService := &mockEmailService{}
	handler := NewVerifyEmailHandler(repo, emailService, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if len(emailService.sentMessages) != 0 {
		t.Errorf("Expected 0 emails sent, got %d", len(emailService.sentMessages))
	}
}

func TestVerifyEmailHandler_UpdateError(t *testing.T) {
	token := "valid-token"
	expiry := time.Now().Add(24 * time.Hour)

	user := entities.NewUser("test@example.com", "hashedPassword")
	user.ID = "user-123"
	user.EmailVerified = false
	user.EmailVerificationToken = &token
	user.EmailVerificationExpiry = &expiry

	repo := &mockVerifyEmailUserRepo{
		users: map[string]*entities.User{
			token: user,
		},
		updateFunc: func(ctx context.Context, u *entities.User) error {
			return errors.ErrNotFound
		},
	}

	handler := NewVerifyEmailHandler(repo, nil, false)

	cmd := VerifyEmailCommand{
		Token: token,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("Handle() expected error but got nil")
	}
}

func (m *mockVerifyEmailUserRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockVerifyEmailUserRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
