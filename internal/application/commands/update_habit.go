package commands

import (
	"context"
	"strings"

	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type UpdateHabitCommand struct {
	HabitID       string
	UserID        string
	Name          string
	Description   string
	CarryOver     bool
	TargetValue   *float64
	SpecificDays  []int
	SpecificDates []int
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
	habit.CarryOver = cmd.CarryOver
	habit.TargetValue = cmd.TargetValue
	habit.SpecificDays = cmd.SpecificDays
	habit.SpecificDates = cmd.SpecificDates

	return h.habitRepo.Update(ctx, habit)
}
