package value_objects

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
