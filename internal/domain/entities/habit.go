package entities

import (
	"time"

	"habit-tracker-api/internal/domain/value_objects"
)

type Habit struct {
	ID            string
	UserID        string
	Name          string
	Description   string
	Type          value_objects.HabitType
	Frequency     value_objects.Frequency
	SpecificDays  []int
	SpecificDates []int
	CarryOver     bool
	TargetValue   *float64
	CreatedAt     time.Time
	ArchivedAt    *time.Time
}

func NewHabit(
	userID string,
	name string,
	habitType value_objects.HabitType,
	frequency value_objects.Frequency,
	carryOver bool,
) *Habit {
	return &Habit{
		UserID:    userID,
		Name:      name,
		Type:      habitType,
		Frequency: frequency,
		CarryOver: carryOver,
		CreatedAt: time.Now(),
	}
}

func (h *Habit) Archive() {
	now := time.Now()
	h.ArchivedAt = &now
}

func (h *Habit) IsActive() bool {
	return h.ArchivedAt == nil
}
