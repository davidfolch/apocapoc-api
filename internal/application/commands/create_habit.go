package commands

import (
	"context"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type CreateHabitCommand struct {
	UserID        string
	Name          string
	Description   string
	Type          string
	Frequency     string
	SpecificDays  []int
	SpecificDates []int
	CarryOver     bool
	TargetValue   *float64
}

type CreateHabitHandler struct {
	habitRepo repositories.HabitRepository
}

func NewCreateHabitHandler(habitRepo repositories.HabitRepository) *CreateHabitHandler {
	return &CreateHabitHandler{habitRepo: habitRepo}
}

func (h *CreateHabitHandler) Handle(ctx context.Context, cmd CreateHabitCommand) (string, error) {
	habitType := value_objects.HabitType(cmd.Type)
	if !habitType.IsValid() {
		return "", errors.ErrInvalidInput
	}

	frequency := value_objects.Frequency(cmd.Frequency)
	if !frequency.IsValid() {
		return "", errors.ErrInvalidInput
	}

	if frequency == value_objects.FrequencyWeekly && len(cmd.SpecificDays) == 0 {
		return "", errors.ErrInvalidInput
	}

	if frequency == value_objects.FrequencyMonthly && len(cmd.SpecificDates) == 0 {
		return "", errors.ErrInvalidInput
	}

	habit := entities.NewHabit(cmd.UserID, cmd.Name, habitType, frequency, cmd.CarryOver)
	habit.Description = cmd.Description
	habit.SpecificDays = cmd.SpecificDays
	habit.SpecificDates = cmd.SpecificDates
	habit.TargetValue = cmd.TargetValue

	if err := h.habitRepo.Create(ctx, habit); err != nil {
		return "", err
	}

	return habit.ID, nil
}
