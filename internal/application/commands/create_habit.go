package commands

import (
	"context"
	"fmt"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type CreateHabitCommand struct {
	UserID        string
	Name          string
	Description   string
	Type          value_objects.HabitType
	Frequency     value_objects.Frequency
	SpecificDays  []int
	SpecificDates []int
	CarryOver     bool
	IsNegative    bool
	TargetValue   *float64
}

type CreateHabitHandler struct {
	habitRepo repositories.HabitRepository
}

func NewCreateHabitHandler(habitRepo repositories.HabitRepository) *CreateHabitHandler {
	return &CreateHabitHandler{habitRepo: habitRepo}
}

func (h *CreateHabitHandler) Handle(ctx context.Context, cmd CreateHabitCommand) (string, error) {
	if !cmd.Type.IsValid() {
		return "", fmt.Errorf("%w: type: type_invalid", errors.ErrInvalidInput)
	}

	if !cmd.Frequency.IsValid() {
		return "", fmt.Errorf("%w: frequency: frequency_invalid", errors.ErrInvalidInput)
	}

	if cmd.Frequency == value_objects.FrequencyWeekly && len(cmd.SpecificDays) == 0 {
		return "", fmt.Errorf("%w: specific_days: specific_days_required", errors.ErrInvalidInput)
	}

	if cmd.Frequency == value_objects.FrequencyMonthly && len(cmd.SpecificDates) == 0 {
		return "", fmt.Errorf("%w: specific_dates: specific_dates_required", errors.ErrInvalidInput)
	}

	habit := entities.NewHabit(cmd.UserID, cmd.Name, cmd.Type, cmd.Frequency, cmd.CarryOver, cmd.IsNegative)
	habit.Description = cmd.Description
	habit.SpecificDays = cmd.SpecificDays
	habit.SpecificDates = cmd.SpecificDates
	habit.TargetValue = cmd.TargetValue

	if err := h.habitRepo.Create(ctx, habit); err != nil {
		return "", err
	}

	return habit.ID, nil
}
