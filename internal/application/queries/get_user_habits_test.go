package queries

import (
	"apocapoc-api/internal/domain/repositories"
	"context"
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/pagination"
)

type mockGetUserHabitsRepo struct {
	habits []*entities.Habit
}

func (m *mockGetUserHabitsRepo) Create(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockGetUserHabitsRepo) FindByID(ctx context.Context, id string) (*entities.Habit, error) {
	return nil, nil
}

func (m *mockGetUserHabitsRepo) FindByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return nil, nil
}

func (m *mockGetUserHabitsRepo) FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Habit, error) {
	return m.habits, nil
}

func (m *mockGetUserHabitsRepo) FindActiveByUserIDWithPagination(ctx context.Context, userID string, params pagination.Params) ([]*entities.Habit, error) {
	offset := params.Offset()
	limit := params.Limit()

	if offset >= len(m.habits) {
		return []*entities.Habit{}, nil
	}

	end := offset + limit
	if end > len(m.habits) {
		end = len(m.habits)
	}

	return m.habits[offset:end], nil
}

func (m *mockGetUserHabitsRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	return len(m.habits), nil
}

func (m *mockGetUserHabitsRepo) Update(ctx context.Context, habit *entities.Habit) error {
	return nil
}

func (m *mockGetUserHabitsRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func TestGetUserHabitsHandler_ReturnsAllActiveHabits(t *testing.T) {
	habit1 := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit1.ID = "habit-1"

	habit2 := entities.NewHabit("user-123", "Read", value_objects.HabitTypeBoolean, value_objects.FrequencyWeekly, false, false)
	habit2.ID = "habit-2"

	habitRepo := &mockGetUserHabitsRepo{habits: []*entities.Habit{habit1, habit2}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-123",
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Habits) != 2 {
		t.Fatalf("Expected 2 habits, got %d", len(result.Habits))
	}

	if result.Habits[0].ID != "habit-1" {
		t.Errorf("Expected first habit ID habit-1, got %s", result.Habits[0].ID)
	}

	if result.Habits[1].ID != "habit-2" {
		t.Errorf("Expected second habit ID habit-2, got %s", result.Habits[1].ID)
	}

	if result.Pagination != nil {
		t.Error("Expected no pagination when not requested")
	}
}

func TestGetUserHabitsHandler_ReturnsEmptyListForUserWithNoHabits(t *testing.T) {
	habitRepo := &mockGetUserHabitsRepo{habits: []*entities.Habit{}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-456",
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Habits) != 0 {
		t.Fatalf("Expected 0 habits, got %d", len(result.Habits))
	}
}

func TestGetUserHabitsHandler_IncludesAllHabitFields(t *testing.T) {
	targetValue := 5.0
	habit := entities.NewHabit("user-123", "Drink Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, true, false)
	habit.ID = "habit-1"
	habit.TargetValue = &targetValue

	habitRepo := &mockGetUserHabitsRepo{habits: []*entities.Habit{habit}}

	handler := NewGetUserHabitsHandler(habitRepo)

	query := GetUserHabitsQuery{
		UserID: "user-123",
	}

	result, err := handler.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Habits) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(result.Habits))
	}

	habitDTO := result.Habits[0]

	if habitDTO.Name != "Drink Water" {
		t.Errorf("Expected name 'Drink Water', got %s", habitDTO.Name)
	}

	if habitDTO.Type != value_objects.HabitTypeValue {
		t.Errorf("Expected type %s, got %s", value_objects.HabitTypeValue, habitDTO.Type)
	}

	if habitDTO.Frequency != value_objects.FrequencyDaily {
		t.Errorf("Expected frequency %s, got %s", value_objects.FrequencyDaily, habitDTO.Frequency)
	}

	if habitDTO.TargetValue == nil || *habitDTO.TargetValue != 5.0 {
		t.Errorf("Expected target value 5.0, got %v", habitDTO.TargetValue)
	}

	if !habitDTO.CarryOver {
		t.Error("Expected carry over to be true")
	}
}

func TestGetUserHabitsHandler_WithPagination(t *testing.T) {
	var habits []*entities.Habit
	for i := 1; i <= 10; i++ {
		habit := entities.NewHabit("user-123", "Habit", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
		habit.ID = "habit-" + string(rune(i+'0'))
		habits = append(habits, habit)
	}

	habitRepo := &mockGetUserHabitsRepo{habits: habits}
	handler := NewGetUserHabitsHandler(habitRepo)

	t.Run("FirstPage", func(t *testing.T) {
		params := pagination.NewParams(1, 5)
		query := GetUserHabitsQuery{
			UserID:           "user-123",
			PaginationParams: &params,
		}

		result, err := handler.Handle(context.Background(), query)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 5 {
			t.Errorf("Expected 5 habits on first page, got %d", len(result.Habits))
		}

		if result.Pagination == nil {
			t.Fatal("Expected pagination metadata")
		}

		if result.Pagination.Page != 1 {
			t.Errorf("Expected page 1, got %d", result.Pagination.Page)
		}

		if result.Pagination.PageSize != 5 {
			t.Errorf("Expected page_size 5, got %d", result.Pagination.PageSize)
		}

		if result.Pagination.TotalItems != 10 {
			t.Errorf("Expected total_items 10, got %d", result.Pagination.TotalItems)
		}

		if result.Pagination.TotalPages != 2 {
			t.Errorf("Expected total_pages 2, got %d", result.Pagination.TotalPages)
		}
	})

	t.Run("SecondPage", func(t *testing.T) {
		params := pagination.NewParams(2, 5)
		query := GetUserHabitsQuery{
			UserID:           "user-123",
			PaginationParams: &params,
		}

		result, err := handler.Handle(context.Background(), query)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 5 {
			t.Errorf("Expected 5 habits on second page, got %d", len(result.Habits))
		}

		if result.Pagination.Page != 2 {
			t.Errorf("Expected page 2, got %d", result.Pagination.Page)
		}
	})

	t.Run("PageBeyondTotal", func(t *testing.T) {
		params := pagination.NewParams(10, 5)
		query := GetUserHabitsQuery{
			UserID:           "user-123",
			PaginationParams: &params,
		}

		result, err := handler.Handle(context.Background(), query)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 0 {
			t.Errorf("Expected 0 habits beyond total, got %d", len(result.Habits))
		}
	})

	t.Run("CustomPageSize", func(t *testing.T) {
		params := pagination.NewParams(1, 3)
		query := GetUserHabitsQuery{
			UserID:           "user-123",
			PaginationParams: &params,
		}

		result, err := handler.Handle(context.Background(), query)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 3 {
			t.Errorf("Expected 3 habits with page_size=3, got %d", len(result.Habits))
		}

		if result.Pagination.PageSize != 3 {
			t.Errorf("Expected page_size 3, got %d", result.Pagination.PageSize)
		}

		if result.Pagination.TotalPages != 4 {
			t.Errorf("Expected total_pages 4 (10 items / 3 per page), got %d", result.Pagination.TotalPages)
		}
	})
}

func (m *mockGetUserHabitsRepo) FindByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter, paginationParams *pagination.Params) ([]*entities.Habit, error) {
	var filtered []*entities.Habit

	for _, habit := range m.habits {
		if filter.Type != nil && habit.Type != *filter.Type {
			continue
		}
		if filter.Frequency != nil && habit.Frequency != *filter.Frequency {
			continue
		}
		if !filter.IncludeArchived && habit.ArchivedAt != nil {
			continue
		}
		filtered = append(filtered, habit)
	}

	if paginationParams != nil {
		offset := paginationParams.Offset()
		limit := paginationParams.Limit()

		if offset >= len(filtered) {
			return []*entities.Habit{}, nil
		}

		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}

		return filtered[offset:end], nil
	}

	return filtered, nil
}

func (m *mockGetUserHabitsRepo) CountByUserIDFiltered(ctx context.Context, userID string, filter repositories.HabitFilter) (int, error) {
	count := 0

	for _, habit := range m.habits {
		if filter.Type != nil && habit.Type != *filter.Type {
			continue
		}
		if filter.Frequency != nil && habit.Frequency != *filter.Frequency {
			continue
		}
		if !filter.IncludeArchived && habit.ArchivedAt != nil {
			continue
		}
		count++
	}

	return count, nil
}

func (m *mockGetUserHabitsRepo) GetChangesSince(ctx context.Context, userID string, since time.Time) (*repositories.HabitChanges, error) {
	return &repositories.HabitChanges{
		Created: []*entities.Habit{},
		Updated: []*entities.Habit{},
		Deleted: []string{},
	}, nil
}

func (m *mockGetUserHabitsRepo) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func TestGetUserHabitsHandler_WithFilters(t *testing.T) {
	habit1 := entities.NewHabit("user-123", "Exercise", value_objects.HabitTypeBoolean, value_objects.FrequencyDaily, false, false)
	habit1.ID = "habit-1"

	habit2 := entities.NewHabit("user-123", "Read", value_objects.HabitTypeCounter, value_objects.FrequencyWeekly, false, false)
	habit2.ID = "habit-2"

	habit3 := entities.NewHabit("user-123", "Water", value_objects.HabitTypeValue, value_objects.FrequencyDaily, false, false)
	habit3.ID = "habit-3"

	habitRepo := &mockGetUserHabitsRepo{habits: []*entities.Habit{habit1, habit2, habit3}}
	handler := NewGetUserHabitsHandler(habitRepo)

	t.Run("FilterByType", func(t *testing.T) {
		habitType := value_objects.HabitTypeBoolean
		query := GetUserHabitsQuery{
			UserID: "user-123",
			FilterParams: &FilterParams{
				Type: &habitType,
			},
		}

		result, err := handler.Handle(context.Background(), query)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 1 {
			t.Errorf("Expected 1 BOOLEAN habit, got %d", len(result.Habits))
		}

		if result.Habits[0].Type != value_objects.HabitTypeBoolean {
			t.Errorf("Expected BOOLEAN type, got %s", result.Habits[0].Type)
		}
	})

	t.Run("FilterByFrequency", func(t *testing.T) {
		frequency := value_objects.FrequencyDaily
		query := GetUserHabitsQuery{
			UserID: "user-123",
			FilterParams: &FilterParams{
				Frequency: &frequency,
			},
		}

		result, err := handler.Handle(context.Background(), query)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 2 {
			t.Errorf("Expected 2 DAILY habits, got %d", len(result.Habits))
		}
	})

	t.Run("FilterWithPagination", func(t *testing.T) {
		frequency := value_objects.FrequencyDaily
		params := pagination.NewParams(1, 1)
		query := GetUserHabitsQuery{
			UserID: "user-123",
			FilterParams: &FilterParams{
				Frequency: &frequency,
			},
			PaginationParams: &params,
		}

		result, err := handler.Handle(context.Background(), query)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Habits) != 1 {
			t.Errorf("Expected 1 habit on first page, got %d", len(result.Habits))
		}

		if result.Pagination == nil {
			t.Fatal("Expected pagination metadata")
		}

		if result.Pagination.TotalItems != 2 {
			t.Errorf("Expected 2 total DAILY habits, got %d", result.Pagination.TotalItems)
		}
	})
}
