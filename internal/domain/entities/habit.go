package entities

import (
	"time"

	"apocapoc-api/internal/domain/value_objects"
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
	IsNegative    bool
	TargetValue   *float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ArchivedAt    *time.Time
	DeletedAt     *time.Time
}

func NewHabit(
	userID string,
	name string,
	habitType value_objects.HabitType,
	frequency value_objects.Frequency,
	carryOver bool,
	isNegative bool,
) *Habit {
	now := time.Now()
	return &Habit{
		UserID:     userID,
		Name:       name,
		Type:       habitType,
		Frequency:  frequency,
		CarryOver:  carryOver,
		IsNegative: isNegative,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (h *Habit) Archive() {
	now := time.Now()
	h.ArchivedAt = &now
	h.UpdatedAt = now
}

func (h *Habit) IsActive() bool {
	return h.ArchivedAt == nil
}

func (h *Habit) Delete() {
	now := time.Now()
	h.DeletedAt = &now
	h.UpdatedAt = now
}

func (h *Habit) IsDeleted() bool {
	return h.DeletedAt != nil
}

func (h *Habit) Touch() {
	h.UpdatedAt = time.Now()
}
