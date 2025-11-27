package value_objects

import (
	"encoding/json"
	"testing"
)

func TestHabitType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		habitType HabitType
		expected  string
	}{
		{"Boolean", HabitTypeBoolean, `"BOOLEAN"`},
		{"Counter", HabitTypeCounter, `"COUNTER"`},
		{"Value", HabitTypeValue, `"VALUE"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.habitType)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestHabitType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  HabitType
		shouldErr bool
	}{
		{"Valid Boolean", `"BOOLEAN"`, HabitTypeBoolean, false},
		{"Valid Counter", `"COUNTER"`, HabitTypeCounter, false},
		{"Valid Value", `"VALUE"`, HabitTypeValue, false},
		{"Invalid type", `"INVALID"`, "", true},
		{"Empty string", `""`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ht HabitType
			err := json.Unmarshal([]byte(tt.input), &ht)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if ht != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, ht)
			}
		})
	}
}

func TestHabitType_JSONRoundTrip(t *testing.T) {
	type testStruct struct {
		Type HabitType `json:"type"`
	}

	original := testStruct{Type: HabitTypeCounter}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded testStruct
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("Round trip failed: expected %s, got %s", original.Type, decoded.Type)
	}
}
