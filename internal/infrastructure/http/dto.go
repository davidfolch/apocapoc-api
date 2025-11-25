package http

import "time"

type CreateHabitRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	Frequency     string   `json:"frequency"`
	SpecificDays  []int    `json:"specific_days,omitempty"`
	SpecificDates []int    `json:"specific_dates,omitempty"`
	CarryOver     bool     `json:"carry_over"`
	TargetValue   *float64 `json:"target_value,omitempty"`
}

type HabitResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Type          string     `json:"type"`
	Frequency     string     `json:"frequency"`
	SpecificDays  []int      `json:"specific_days,omitempty"`
	SpecificDates []int      `json:"specific_dates,omitempty"`
	CarryOver     bool       `json:"carry_over"`
	TargetValue   *float64   `json:"target_value,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ArchivedAt    *time.Time `json:"archived_at,omitempty"`
}

type MarkHabitRequest struct {
	ScheduledDate string   `json:"scheduled_date"`
	Value         *float64 `json:"value,omitempty"`
}

type TodaysHabitResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	TargetValue   *float64  `json:"target_value,omitempty"`
	ScheduledDate time.Time `json:"scheduled_date"`
	IsCarriedOver bool      `json:"is_carried_over"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
