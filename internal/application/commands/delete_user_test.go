package commands

import (
	"apocapoc-api/internal/shared/pagination"
	"context"
	"testing"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/errors"
)

type mockDeleteUserRepo struct {
	findByIDFunc func(ctx context.Context, id string) (*entities.User, error)
	deleteFunc   func(ctx context.Context, id string) error
}

func (m *mockDeleteUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.ErrNotFound
}

func (m *mockDeleteUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockDeleteUserRepo) FindByVerificationToken(ctx context.Context, token string) (*entities.User, error) {
	return nil, errors.ErrNotFound
}

func (m *mockDeleteUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockDeleteUserRepo) Update(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockDeleteUserRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestDeleteUserHandler_Success(t *testing.T) {
	var deletedID string

	repo := &mockDeleteUserRepo{
		findByIDFunc: func(ctx context.Context, id string) (*entities.User, error) {
			user := entities.NewUser("test@example.com", "hashedPassword")
			user.ID = id
			return user, nil
		},
		deleteFunc: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}

	handler := NewDeleteUserHandler(repo)

	cmd := DeleteUserCommand{
		UserID: "user-123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	if deletedID != "user-123" {
		t.Errorf("deletedID = %v, want %v", deletedID, "user-123")
	}
}

func TestDeleteUserHandler_EmptyUserID(t *testing.T) {
	repo := &mockDeleteUserRepo{}
	handler := NewDeleteUserHandler(repo)

	cmd := DeleteUserCommand{
		UserID: "",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrInvalidInput {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrInvalidInput)
	}
}

func TestDeleteUserHandler_UserNotFound(t *testing.T) {
	repo := &mockDeleteUserRepo{
		findByIDFunc: func(ctx context.Context, id string) (*entities.User, error) {
			return nil, errors.ErrNotFound
		},
	}

	handler := NewDeleteUserHandler(repo)

	cmd := DeleteUserCommand{
		UserID: "non-existent-user",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != errors.ErrNotFound {
		t.Errorf("Handle() error = %v, want %v", err, errors.ErrNotFound)
	}
}

func TestDeleteUserHandler_DeleteError(t *testing.T) {
	customError := errors.ErrNotFound

	repo := &mockDeleteUserRepo{
		findByIDFunc: func(ctx context.Context, id string) (*entities.User, error) {
			user := entities.NewUser("test@example.com", "hashedPassword")
			user.ID = id
			return user, nil
		},
		deleteFunc: func(ctx context.Context, id string) error {
			return customError
		},
	}

	handler := NewDeleteUserHandler(repo)

	cmd := DeleteUserCommand{
		UserID: "user-123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != customError {
		t.Errorf("Handle() error = %v, want %v", err, customError)
	}
}

func (m *mockDeleteUserRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockDeleteUserRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
