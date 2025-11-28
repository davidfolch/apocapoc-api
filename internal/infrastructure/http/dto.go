package http

import (
	"time"

	"apocapoc-api/internal/domain/value_objects"
)

type CreateHabitRequest struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Type          value_objects.HabitType `json:"type"`
	Frequency     value_objects.Frequency `json:"frequency"`
	SpecificDays  []int                   `json:"specific_days,omitempty"`
	SpecificDates []int                   `json:"specific_dates,omitempty"`
	CarryOver     bool                    `json:"carry_over"`
	IsNegative    bool                    `json:"is_negative"`
	TargetValue   *float64                `json:"target_value,omitempty"`
}

type UpdateHabitRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SpecificDays  []int    `json:"specific_days,omitempty"`
	SpecificDates []int    `json:"specific_dates,omitempty"`
	CarryOver     bool     `json:"carry_over"`
	TargetValue   *float64 `json:"target_value,omitempty"`
}

type HabitResponse struct {
	ID            string                  `json:"id"`
	UserID        string                  `json:"user_id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Type          value_objects.HabitType `json:"type"`
	Frequency     value_objects.Frequency `json:"frequency"`
	SpecificDays  []int                   `json:"specific_days,omitempty"`
	SpecificDates []int                   `json:"specific_dates,omitempty"`
	CarryOver     bool                    `json:"carry_over"`
	IsNegative    bool                    `json:"is_negative"`
	TargetValue   *float64                `json:"target_value,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	ArchivedAt    *time.Time              `json:"archived_at,omitempty"`
}

type MarkHabitRequest struct {
	ScheduledDate string   `json:"scheduled_date"`
	Value         *float64 `json:"value,omitempty"`
}

type TodaysHabitResponse struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Type          value_objects.HabitType `json:"type"`
	TargetValue   *float64                `json:"target_value,omitempty"`
	IsNegative    bool                    `json:"is_negative"`
	ScheduledDate time.Time               `json:"scheduled_date"`
	IsCarriedOver bool                    `json:"is_carried_over"`
}

type UserHabitResponse struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Type         value_objects.HabitType `json:"type"`
	Frequency    value_objects.Frequency `json:"frequency"`
	SpecificDays []int                   `json:"specific_days,omitempty"`
	TargetValue  *float64                `json:"target_value,omitempty"`
	CarryOver    bool                    `json:"carry_over"`
	IsNegative   bool                    `json:"is_negative"`
}

type HabitEntryResponse struct {
	ID            string    `json:"id"`
	HabitID       string    `json:"habit_id"`
	ScheduledDate time.Time `json:"scheduled_date"`
	CompletedAt   time.Time `json:"completed_at"`
	Value         *float64  `json:"value,omitempty"`
}

type HabitEntriesResponse struct {
	Entries []HabitEntryResponse `json:"entries"`
	Total   int                  `json:"total"`
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error string `json:"error"`
	Field string `json:"field"`
}
