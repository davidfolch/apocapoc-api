package value_objects

import "testing"

func TestHabitType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		habType  HabitType
		expected bool
	}{
		{"Boolean type is valid", HabitTypeBoolean, true},
		{"Counter type is valid", HabitTypeCounter, true},
		{"Value type is valid", HabitTypeValue, true},
		{"Empty string is invalid", HabitType(""), false},
		{"Random string is invalid", HabitType("RANDOM"), false},
		{"Lowercase is invalid", HabitType("boolean"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.habType.IsValid(); got != tt.expected {
				t.Errorf("HabitType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
