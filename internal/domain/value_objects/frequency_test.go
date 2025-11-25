package value_objects

import "testing"

func TestFrequency_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		freq     Frequency
		expected bool
	}{
		{"Daily frequency is valid", FrequencyDaily, true},
		{"Weekly frequency is valid", FrequencyWeekly, true},
		{"Monthly frequency is valid", FrequencyMonthly, true},
		{"Empty string is invalid", Frequency(""), false},
		{"Random string is invalid", Frequency("YEARLY"), false},
		{"Lowercase is invalid", Frequency("daily"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.freq.IsValid(); got != tt.expected {
				t.Errorf("Frequency.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
