package queries

import (
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockRefreshTokenRepository struct {
	findByTokenFunc func(ctx context.Context, token string) (*entities.RefreshToken, error)
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	return nil
}

func (m *mockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, errors.ErrNotFound
}

func (m *mockRefreshTokenRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	return nil
}

func (m *mockRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	return nil
}

func (m *mockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockUserRepositoryForRefresh struct {
	findByIDFunc func(ctx context.Context, id string) (*entities.User, error)
}

func (m *mockUserRepositoryForRefresh) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockUserRepositoryForRefresh) FindByID(ctx context.Context, id string) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.ErrNotFound
}

func (m *mockUserRepositoryForRefresh) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForRefresh) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockUserRepositoryForRefresh) Update(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockUserRepositoryForRefresh) Delete(ctx context.Context, id string) error {
	return nil
}

func TestRefreshTokenHandler_Success(t *testing.T) {
	refreshTokenRepo := &mockRefreshTokenRepository{
		findByTokenFunc: func(ctx context.Context, token string) (*entities.RefreshToken, error) {
			return entities.NewRefreshToken("user-123", token, time.Now().Add(24*time.Hour)), nil
		},
	}

	userRepo := &mockUserRepositoryForRefresh{
		findByIDFunc: func(ctx context.Context, id string) (*entities.User, error) {
			user := entities.NewUser("test@example.com", "hash")
			user.ID = id
			return user, nil
		},
	}

	handler := NewRefreshTokenHandler(refreshTokenRepo, userRepo)

	query := RefreshTokenQuery{
		RefreshToken: "valid-token",
	}

	result, err := handler.Handle(context.Background(), query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.UserID != "user-123" {
		t.Errorf("Expected userID 'user-123', got %s", result.UserID)
	}

	if result.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", result.Email)
	}
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	refreshTokenRepo := &mockRefreshTokenRepository{}
	userRepo := &mockUserRepositoryForRefresh{}

	handler := NewRefreshTokenHandler(refreshTokenRepo, userRepo)

	query := RefreshTokenQuery{
		RefreshToken: "invalid-token",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestRefreshTokenHandler_ExpiredToken(t *testing.T) {
	refreshTokenRepo := &mockRefreshTokenRepository{
		findByTokenFunc: func(ctx context.Context, token string) (*entities.RefreshToken, error) {
			return entities.NewRefreshToken("user-123", token, time.Now().Add(-1*time.Hour)), nil
		},
	}

	userRepo := &mockUserRepositoryForRefresh{}

	handler := NewRefreshTokenHandler(refreshTokenRepo, userRepo)

	query := RefreshTokenQuery{
		RefreshToken: "expired-token",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound for expired token, got %v", err)
	}
}

func TestRefreshTokenHandler_RevokedToken(t *testing.T) {
	refreshToken := entities.NewRefreshToken("user-123", "revoked-token", time.Now().Add(24*time.Hour))
	refreshToken.Revoke()

	refreshTokenRepo := &mockRefreshTokenRepository{
		findByTokenFunc: func(ctx context.Context, token string) (*entities.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	userRepo := &mockUserRepositoryForRefresh{}

	handler := NewRefreshTokenHandler(refreshTokenRepo, userRepo)

	query := RefreshTokenQuery{
		RefreshToken: "revoked-token",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrNotFound {
		t.Errorf("Expected ErrNotFound for revoked token, got %v", err)
	}
}

func TestRefreshTokenHandler_EmptyToken(t *testing.T) {
	refreshTokenRepo := &mockRefreshTokenRepository{}
	userRepo := &mockUserRepositoryForRefresh{}

	handler := NewRefreshTokenHandler(refreshTokenRepo, userRepo)

	query := RefreshTokenQuery{
		RefreshToken: "",
	}

	_, err := handler.Handle(context.Background(), query)
	if err != errors.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() unexpected error = %v", err)
	}

	if token1 == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}

	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() unexpected error = %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateRefreshToken() generated identical tokens")
	}

	if len(token1) < 20 {
		t.Errorf("GenerateRefreshToken() token too short: %d characters", len(token1))
	}
}

func TestCreateRefreshToken(t *testing.T) {
	userID := "user-123"
	expiryDuration := 7 * 24 * time.Hour

	token, err := CreateRefreshToken(userID, expiryDuration)
	if err != nil {
		t.Fatalf("CreateRefreshToken() unexpected error = %v", err)
	}

	if token == nil {
		t.Fatal("CreateRefreshToken() returned nil")
	}

	if token.UserID != userID {
		t.Errorf("UserID = %v, want %v", token.UserID, userID)
	}

	if token.Token == "" {
		t.Error("Token is empty")
	}

	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt is zero")
	}

	if token.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	expectedExpiry := time.Now().Add(expiryDuration)
	diff := token.ExpiresAt.Sub(expectedExpiry)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("ExpiresAt difference too large: %v", diff)
	}

	if !token.IsValid() {
		t.Error("Token should be valid")
	}
}

func TestCreateRefreshToken_MultipleCalls(t *testing.T) {
	token1, err := CreateRefreshToken("user-1", 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRefreshToken(1) unexpected error = %v", err)
	}

	token2, err := CreateRefreshToken("user-2", 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateRefreshToken(2) unexpected error = %v", err)
	}

	if token1.Token == token2.Token {
		t.Error("CreateRefreshToken() generated identical tokens for different users")
	}

	if token1.UserID == token2.UserID {
		t.Error("UserIDs should be different")
	}
}

func (m *mockRefreshTokenRepository) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockUserRepositoryForRefresh) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockUserRepositoryForRefresh) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
