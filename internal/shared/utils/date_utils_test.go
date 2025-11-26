package utils

import (
	"testing"
	"time"
)

func TestShouldAppearToday_Daily(t *testing.T) {
	result := ShouldAppearToday("DAILY", nil, nil, time.Now())
	if !result {
		t.Error("Daily habit should appear every day")
	}
}

func TestShouldAppearToday_Weekly(t *testing.T) {
	tests := []struct {
		name         string
		specificDays []int
		targetDate   time.Time
		expected     bool
	}{
		{
			name:         "Monday when Monday is specified",
			specificDays: []int{1}, // Monday
			targetDate:   time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), // Monday
			expected:     true,
		},
		{
			name:         "Tuesday when Monday is specified",
			specificDays: []int{1}, // Monday
			targetDate:   time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC), // Tuesday
			expected:     false,
		},
		{
			name:         "Wednesday when Mon/Wed/Fri specified",
			specificDays: []int{1, 3, 5}, // Mon, Wed, Fri
			targetDate:   time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC), // Wednesday
			expected:     true,
		},
		{
			name:         "Sunday when Mon/Wed/Fri specified",
			specificDays: []int{1, 3, 5}, // Mon, Wed, Fri
			targetDate:   time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), // Sunday
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldAppearToday("WEEKLY", tt.specificDays, nil, tt.targetDate)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %s", tt.expected, result, tt.targetDate.Weekday())
			}
		})
	}
}

func TestShouldAppearToday_Monthly(t *testing.T) {
	tests := []struct {
		name          string
		specificDates []int
		targetDate    time.Time
		expected      bool
	}{
		{
			name:          "1st of month when 1st is specified",
			specificDates: []int{1},
			targetDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:      true,
		},
		{
			name:          "2nd of month when 1st is specified",
			specificDates: []int{1},
			targetDate:    time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			expected:      false,
		},
		{
			name:          "15th when 1st/15th/30th specified",
			specificDates: []int{1, 15, 30},
			targetDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expected:      true,
		},
		{
			name:          "10th when 1st/15th/30th specified",
			specificDates: []int{1, 15, 30},
			targetDate:    time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldAppearToday("MONTHLY", nil, tt.specificDates, tt.targetDate)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for day %d", tt.expected, result, tt.targetDate.Day())
			}
		})
	}
}

func TestShouldAppearToday_InvalidFrequency(t *testing.T) {
	result := ShouldAppearToday("INVALID", nil, nil, time.Now())
	if result {
		t.Error("Invalid frequency should return false")
	}
}
