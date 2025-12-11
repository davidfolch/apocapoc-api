package entities

import "time"

type HabitEntry struct {
	ID            string
	HabitID       string
	ScheduledDate time.Time
	CompletedAt   time.Time
	Value         *float64
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func NewHabitEntry(habitID string, scheduledDate time.Time, value *float64) *HabitEntry {
	now := time.Now()
	return &HabitEntry{
		HabitID:       habitID,
		ScheduledDate: scheduledDate,
		CompletedAt:   now,
		Value:         value,
		UpdatedAt:     now,
	}
}

func (e *HabitEntry) Delete() {
	now := time.Now()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

func (e *HabitEntry) IsDeleted() bool {
	return e.DeletedAt != nil
}
