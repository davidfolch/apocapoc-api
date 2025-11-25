package entities

import "time"

type HabitEntry struct {
	ID            string
	HabitID       string
	ScheduledDate time.Time
	CompletedAt   time.Time
	Value         *float64
	DeletedAt     *time.Time
}

func NewHabitEntry(habitID string, scheduledDate time.Time, value *float64) *HabitEntry {
	return &HabitEntry{
		HabitID:       habitID,
		ScheduledDate: scheduledDate,
		CompletedAt:   time.Now(),
		Value:         value,
	}
}

func (e *HabitEntry) Delete() {
	now := time.Now()
	e.DeletedAt = &now
}

func (e *HabitEntry) IsDeleted() bool {
	return e.DeletedAt != nil
}
