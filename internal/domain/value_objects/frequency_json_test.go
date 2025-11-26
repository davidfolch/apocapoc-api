package value_objects

import (
	"encoding/json"
	"testing"
)

func TestFrequency_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		frequency Frequency
		expected  string
	}{
		{"Daily", FrequencyDaily, `"DAILY"`},
		{"Weekly", FrequencyWeekly, `"WEEKLY"`},
		{"Monthly", FrequencyMonthly, `"MONTHLY"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.frequency)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestFrequency_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Frequency
		shouldErr bool
	}{
		{"Valid Daily", `"DAILY"`, FrequencyDaily, false},
		{"Valid Weekly", `"WEEKLY"`, FrequencyWeekly, false},
		{"Valid Monthly", `"MONTHLY"`, FrequencyMonthly, false},
		{"Invalid frequency", `"YEARLY"`, "", true},
		{"Empty string", `""`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Frequency
			err := json.Unmarshal([]byte(tt.input), &f)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if f != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, f)
			}
		})
	}
}

func TestFrequency_JSONRoundTrip(t *testing.T) {
	type testStruct struct {
		Frequency Frequency `json:"frequency"`
	}

	original := testStruct{Frequency: FrequencyWeekly}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded testStruct
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Frequency != original.Frequency {
		t.Errorf("Round trip failed: expected %s, got %s", original.Frequency, decoded.Frequency)
	}
}
