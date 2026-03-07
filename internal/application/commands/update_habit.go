package commands

import (
	"context"
	"strings"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/value_objects"
	"apocapoc-api/internal/shared/errors"
)

type UpdateHabitCommand struct {
	HabitID       string
	UserID        string
	Name          string
	Description   string
	Frequency     value_objects.Frequency
	SpecificDays  []int
	SpecificDates []int
	CarryOver     bool
	TargetValue   *float64
}

type UpdateHabitHandler struct {
	habitRepo repositories.HabitRepository
}

func NewUpdateHabitHandler(habitRepo repositories.HabitRepository) *UpdateHabitHandler {
	return &UpdateHabitHandler{
		habitRepo: habitRepo,
	}
}

func (h *UpdateHabitHandler) Handle(ctx context.Context, cmd UpdateHabitCommand) error {
	if strings.TrimSpace(cmd.Name) == "" {
		return errors.ErrInvalidInput
	}

	if !cmd.Frequency.IsValid() {
		return errors.ErrInvalidInput
	}

	if cmd.Frequency == value_objects.FrequencyWeekly && len(cmd.SpecificDays) == 0 {
		return errors.ErrInvalidInput
	}

	if cmd.Frequency == value_objects.FrequencyMonthly && len(cmd.SpecificDates) == 0 {
		return errors.ErrInvalidInput
	}

	habit, err := h.habitRepo.FindByID(ctx, cmd.HabitID)
	if err != nil {
		return err
	}

	if habit.UserID != cmd.UserID {
		return errors.ErrUnauthorized
	}

	if !habit.IsActive() {
		return errors.ErrInvalidInput
	}

	habit.Name = cmd.Name
	habit.Description = cmd.Description
	habit.Frequency = cmd.Frequency
	habit.SpecificDays = cmd.SpecificDays
	habit.SpecificDates = cmd.SpecificDates
	habit.CarryOver = cmd.CarryOver
	habit.TargetValue = cmd.TargetValue

	return h.habitRepo.Update(ctx, habit)
}
