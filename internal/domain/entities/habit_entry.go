package entities

import "time"

type HabitEntry struct {
	ID            string
	HabitID       string
	ScheduledDate time.Time
	CompletedAt   time.Time
	Value         *float64
}

func NewHabitEntry(habitID string, scheduledDate time.Time, value *float64) *HabitEntry {
	return &HabitEntry{
		HabitID:       habitID,
		ScheduledDate: scheduledDate,
		CompletedAt:   time.Now(),
		Value:         value,
	}
}
