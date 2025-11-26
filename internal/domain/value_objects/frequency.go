package value_objects

import (
	"encoding/json"
	"fmt"
)

type Frequency string

const (
	FrequencyDaily   Frequency = "DAILY"
	FrequencyWeekly  Frequency = "WEEKLY"
	FrequencyMonthly Frequency = "MONTHLY"
)

func (f Frequency) IsValid() bool {
	switch f {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly:
		return true
	}
	return false
}

func (f Frequency) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f *Frequency) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*f = Frequency(s)
	if !f.IsValid() {
		return fmt.Errorf("invalid frequency: %s (must be DAILY, WEEKLY, or MONTHLY)", s)
	}

	return nil
}
