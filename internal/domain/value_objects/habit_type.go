package value_objects

import (
	"encoding/json"
	"fmt"
)

type HabitType string

const (
	HabitTypeBoolean HabitType = "BOOLEAN"
	HabitTypeCounter HabitType = "COUNTER"
	HabitTypeValue   HabitType = "VALUE"
)

func (ht HabitType) IsValid() bool {
	switch ht {
	case HabitTypeBoolean, HabitTypeCounter, HabitTypeValue:
		return true
	}
	return false
}

func (ht HabitType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ht))
}

func (ht *HabitType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*ht = HabitType(s)
	if !ht.IsValid() {
		return fmt.Errorf("invalid habit type: %s (must be BOOLEAN, COUNTER, or VALUE)", s)
	}

	return nil
}
