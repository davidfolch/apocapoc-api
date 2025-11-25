package value_objects

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
