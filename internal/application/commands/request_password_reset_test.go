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

type mockRequestResetUserRepo struct {
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
	users           map[string]*entities.User
}

func (m *mockRequestResetUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockRequestResetUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockRequestResetUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockRequestResetUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockRequestResetUserRepo) Update(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockRequestResetUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockRequestResetTokenRepo struct {
	createFunc func(ctx context.Context, token *entities.PasswordResetToken) error
	tokens     []*entities.PasswordResetToken
}

func (m *mockRequestResetTokenRepo) Create(ctx context.Context, token *entities.PasswordResetToken) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, token)
	}
	m.tokens = append(m.tokens, token)
	return nil
}

func (m *mockRequestResetTokenRepo) FindByToken(ctx context.Context, token string) (*entities.PasswordResetToken, error) {
	return nil, errors.ErrNotFound
}

func (m *mockRequestResetTokenRepo) Update(ctx context.Context, token *entities.PasswordResetToken) error {
	return nil
}

func (m *mockRequestResetTokenRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockRequestResetEmailService struct {
	sendFunc     func(message services.EmailMessage) error
	sentMessages []services.EmailMessage
}

func (m *mockRequestResetEmailService) Send(message services.EmailMessage) error {
	if m.sendFunc != nil {
		return m.sendFunc(message)
	}
	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func TestRequestPasswordResetHandler_Success(t *testing.T) {
	user := entities.NewUser("test@example.com", "hash")
	user.ID = "user-123"
	user.EmailVerified = true

	userRepo := &mockRequestResetUserRepo{
		users: map[string]*entities.User{
			user.Email: user,
		},
	}

	tokenRepo := &mockRequestResetTokenRepo{
		tokens: []*entities.PasswordResetToken{},
	}

	emailService := &mockRequestResetEmailService{
		sentMessages: []services.EmailMessage{},
	}

	handler := NewRequestPasswordResetHandler(userRepo, tokenRepo, emailService, "http://localhost:8080")

	cmd := RequestPasswordResetCommand{
		Email: user.Email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if len(tokenRepo.tokens) != 1 {
		t.Fatalf("Expected 1 token created, got %d", len(tokenRepo.tokens))
	}

	createdToken := tokenRepo.tokens[0]
	if createdToken.UserID != user.ID {
		t.Errorf("Token UserID = %v, want %v", createdToken.UserID, user.ID)
	}

	if createdToken.Token == "" {
		t.Error("Token string is empty")
	}

	if createdToken.ExpiresAt.Before(time.Now()) {
		t.Error("Token already expired")
	}

	expectedExpiry := time.Now().Add(1 * time.Hour)
	diff := createdToken.ExpiresAt.Sub(expectedExpiry)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("Token expiry = %v, expected around %v", createdToken.ExpiresAt, expectedExpiry)
	}

	if len(emailService.sentMessages) != 1 {
		t.Fatalf("Expected 1 email sent, got %d", len(emailService.sentMessages))
	}

	sentEmail := emailService.sentMessages[0]
	if sentEmail.To != user.Email {
		t.Errorf("Email To = %v, want %v", sentEmail.To, user.Email)
	}

	if sentEmail.Subject != "Password Reset Request" {
		t.Errorf("Email Subject = %v, want %v", sentEmail.Subject, "Password Reset Request")
	}

	if !sentEmail.IsHTML {
		t.Error("Email should be HTML")
	}

	if sentEmail.Body == "" {
		t.Error("Email body is empty")
	}
}

func TestRequestPasswordResetHandler_EmptyEmail(t *testing.T) {
	handler := NewRequestPasswordResetHandler(
		&mockRequestResetUserRepo{users: make(map[string]*entities.User)},
		&mockRequestResetTokenRepo{tokens: []*entities.PasswordResetToken{}},
		&mockRequestResetEmailService{},
		"http://localhost:8080",
	)

	cmd := RequestPasswordResetCommand{
		Email: "",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestRequestPasswordResetHandler_UserNotFound(t *testing.T) {
	handler := NewRequestPasswordResetHandler(
		&mockRequestResetUserRepo{users: make(map[string]*entities.User)},
		&mockRequestResetTokenRepo{tokens: []*entities.PasswordResetToken{}},
		&mockRequestResetEmailService{},
		"http://localhost:8080",
	)

	cmd := RequestPasswordResetCommand{
		Email: "nonexistent@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrNotFound {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrNotFound)
	}
}

func TestRequestPasswordResetHandler_EmailNotVerified(t *testing.T) {
	user := entities.NewUser("test@example.com", "hash")
	user.ID = "user-123"
	user.EmailVerified = false

	userRepo := &mockRequestResetUserRepo{
		users: map[string]*entities.User{
			user.Email: user,
		},
	}

	handler := NewRequestPasswordResetHandler(
		userRepo,
		&mockRequestResetTokenRepo{tokens: []*entities.PasswordResetToken{}},
		&mockRequestResetEmailService{},
		"http://localhost:8080",
	)

	cmd := RequestPasswordResetCommand{
		Email: user.Email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrEmailNotVerified {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrEmailNotVerified)
	}
}

func TestRequestPasswordResetHandler_TokenCreationFailure(t *testing.T) {
	user := entities.NewUser("test@example.com", "hash")
	user.ID = "user-123"
	user.EmailVerified = true

	userRepo := &mockRequestResetUserRepo{
		users: map[string]*entities.User{
			user.Email: user,
		},
	}

	tokenRepo := &mockRequestResetTokenRepo{
		createFunc: func(ctx context.Context, token *entities.PasswordResetToken) error {
			return errors.ErrInvalidInput
		},
	}

	emailService := &mockRequestResetEmailService{}

	handler := NewRequestPasswordResetHandler(userRepo, tokenRepo, emailService, "http://localhost:8080")

	cmd := RequestPasswordResetCommand{
		Email: user.Email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("Handle() expected error but got nil")
	}

	if len(emailService.sentMessages) != 0 {
		t.Error("Email should not be sent if token creation fails")
	}
}

func TestRequestPasswordResetHandler_EmailSendFailure(t *testing.T) {
	user := entities.NewUser("test@example.com", "hash")
	user.ID = "user-123"
	user.EmailVerified = true

	userRepo := &mockRequestResetUserRepo{
		users: map[string]*entities.User{
			user.Email: user,
		},
	}

	tokenRepo := &mockRequestResetTokenRepo{
		tokens: []*entities.PasswordResetToken{},
	}

	emailService := &mockRequestResetEmailService{
		sendFunc: func(message services.EmailMessage) error {
			return errors.ErrInvalidInput
		},
	}

	handler := NewRequestPasswordResetHandler(userRepo, tokenRepo, emailService, "http://localhost:8080")

	cmd := RequestPasswordResetCommand{
		Email: user.Email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("Handle() expected error but got nil")
	}

	if len(tokenRepo.tokens) != 1 {
		t.Error("Token should be created even if email sending fails")
	}
}

func TestRequestPasswordResetHandler_ResetLinkFormat(t *testing.T) {
	user := entities.NewUser("test@example.com", "hash")
	user.ID = "user-123"
	user.EmailVerified = true

	userRepo := &mockRequestResetUserRepo{
		users: map[string]*entities.User{
			user.Email: user,
		},
	}

	tokenRepo := &mockRequestResetTokenRepo{
		tokens: []*entities.PasswordResetToken{},
	}

	emailService := &mockRequestResetEmailService{
		sentMessages: []services.EmailMessage{},
	}

	appURL := "https://myapp.com"
	handler := NewRequestPasswordResetHandler(userRepo, tokenRepo, emailService, appURL)

	cmd := RequestPasswordResetCommand{
		Email: user.Email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if len(emailService.sentMessages) != 1 {
		t.Fatal("Expected 1 email sent")
	}

	sentEmail := emailService.sentMessages[0]
	if sentEmail.Body == "" {
		t.Fatal("Email body is empty")
	}
}

func (m *mockRequestResetTokenRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockRequestResetTokenRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockRequestResetUserRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockRequestResetUserRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
