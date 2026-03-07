package services

import (
	"testing"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/value_objects"
)

func makeEntry(d time.Time, value *float64) *entities.HabitEntry {
	return &entities.HabitEntry{ScheduledDate: d, Value: value}
}

func floatPtr(v float64) *float64 { return &v }

func dt(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func habit(t value_objects.HabitType, negative bool, target *float64) *entities.Habit {
	return &entities.Habit{
		Type:        t,
		Frequency:   value_objects.FrequencyDaily,
		IsNegative:  negative,
		TargetValue: target,
		CreatedAt:   dt(2026, 1, 1),
	}
}

// --- IsDaySuccessful ---

func TestIsDaySuccessful_BooleanPositive(t *testing.T) {
	h := habit(value_objects.HabitTypeBoolean, false, nil)
	em := map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), nil)}

	if !IsDaySuccessful("2026-03-01", em, h) {
		t.Error("entry should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be failure")
	}
}

func TestIsDaySuccessful_BooleanNegative(t *testing.T) {
	h := habit(value_objects.HabitTypeBoolean, true, nil)
	em := map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), nil)}

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be success (resisted)")
	}
	if IsDaySuccessful("2026-03-01", em, h) {
		t.Error("entry should be failure")
	}
}

func TestIsDaySuccessful_CounterPositiveNoTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeCounter, false, nil)

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(5))}, h) {
		t.Error("entry should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be failure")
	}
}

func TestIsDaySuccessful_CounterPositiveWithTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeCounter, false, floatPtr(8))

	cases := []struct {
		name    string
		value   *float64
		has     bool
		success bool
	}{
		{"value >= target", floatPtr(10), true, true},
		{"value == target", floatPtr(8), true, true},
		{"value < target", floatPtr(3), true, false},
		{"no entry", nil, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			em := map[string]*entities.HabitEntry{}
			if tc.has {
				em["2026-03-01"] = makeEntry(dt(2026, 3, 1), tc.value)
			}
			if IsDaySuccessful("2026-03-01", em, h) != tc.success {
				t.Errorf("Expected %v", tc.success)
			}
		})
	}
}

func TestIsDaySuccessful_CounterNegativeNoTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeCounter, true, nil)

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(3))}, h) {
		t.Error("entry should be failure")
	}
}

func TestIsDaySuccessful_CounterNegativeWithTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeCounter, true, floatPtr(2))

	cases := []struct {
		name    string
		value   *float64
		has     bool
		success bool
	}{
		{"no entry", nil, false, true},
		{"within limit", floatPtr(1), true, true},
		{"at limit", floatPtr(2), true, true},
		{"over limit", floatPtr(5), true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			em := map[string]*entities.HabitEntry{}
			if tc.has {
				em["2026-03-01"] = makeEntry(dt(2026, 3, 1), tc.value)
			}
			if IsDaySuccessful("2026-03-01", em, h) != tc.success {
				t.Errorf("Expected %v", tc.success)
			}
		})
	}
}

func TestIsDaySuccessful_ValuePositiveNoTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeValue, false, nil)

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(72))}, h) {
		t.Error("entry should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be failure")
	}
}

func TestIsDaySuccessful_ValuePositiveWithTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeValue, false, floatPtr(7))

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(8))}, h) {
		t.Error("value >= target should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(5))}, h) {
		t.Error("value < target should be failure")
	}
}

func TestIsDaySuccessful_ValueNegativeNoTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeValue, true, nil)

	if !IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{}, h) {
		t.Error("no entry should be success")
	}
	if IsDaySuccessful("2026-03-01", map[string]*entities.HabitEntry{"2026-03-01": makeEntry(dt(2026, 3, 1), floatPtr(3))}, h) {
		t.Error("entry should be failure")
	}
}

func TestIsDaySuccessful_ValueNegativeWithTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeValue, true, floatPtr(70))

	cases := []struct {
		name    string
		value   *float64
		has     bool
		success bool
	}{
		{"below target", floatPtr(68), true, true},
		{"above target", floatPtr(75), true, false},
		{"no entry (didnt track)", nil, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			em := map[string]*entities.HabitEntry{}
			if tc.has {
				em["2026-03-01"] = makeEntry(dt(2026, 3, 1), tc.value)
			}
			if IsDaySuccessful("2026-03-01", em, h) != tc.success {
				t.Errorf("Expected %v", tc.success)
			}
		})
	}
}

// --- CalculateStreaks ---

func TestStreaks_DailyBooleanPositive(t *testing.T) {
	h := habit(value_objects.HabitTypeBoolean, false, nil)

	t.Run("3 consecutive days", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 1), nil),
			makeEntry(dt(2026, 3, 2), nil),
			makeEntry(dt(2026, 3, 3), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 3))
		if r.Current != 3 || r.Longest != 3 {
			t.Errorf("Expected current=3 longest=3, got current=%d longest=%d", r.Current, r.Longest)
		}
	})

	t.Run("gap finds longest and current separately", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 1), nil),
			makeEntry(dt(2026, 3, 2), nil),
			makeEntry(dt(2026, 3, 3), nil),
			// gap Mar 4
			makeEntry(dt(2026, 3, 5), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 5))
		if r.Current != 1 || r.Longest != 3 {
			t.Errorf("Expected current=1 longest=3, got current=%d longest=%d", r.Current, r.Longest)
		}
	})

	t.Run("today not completed doesnt break streak", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 1), nil),
			makeEntry(dt(2026, 3, 2), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 3))
		if r.Current != 2 || r.Longest != 2 {
			t.Errorf("Expected current=2 longest=2, got current=%d longest=%d", r.Current, r.Longest)
		}
	})

	t.Run("missed yesterday breaks streak", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 1), nil),
			makeEntry(dt(2026, 3, 2), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 4))
		if r.Current != 0 || r.Longest != 2 {
			t.Errorf("Expected current=0 longest=2, got current=%d longest=%d", r.Current, r.Longest)
		}
	})
}

func TestStreaks_DailyBooleanNegative(t *testing.T) {
	h := habit(value_objects.HabitTypeBoolean, true, nil)
	h.CreatedAt = dt(2026, 3, 1)

	t.Run("3 days no entries is 3 streak", func(t *testing.T) {
		r := CalculateStreaks(nil, h, dt(2026, 3, 3))
		if r.Current != 3 || r.Longest != 3 {
			t.Errorf("Expected current=3 longest=3, got current=%d longest=%d", r.Current, r.Longest)
		}
	})

	t.Run("entry breaks streak", func(t *testing.T) {
		entries := []*entities.HabitEntry{makeEntry(dt(2026, 3, 2), nil)}
		r := CalculateStreaks(entries, h, dt(2026, 3, 3))
		if r.Current != 1 || r.Longest != 1 {
			t.Errorf("Expected current=1 longest=1, got current=%d longest=%d", r.Current, r.Longest)
		}
	})
}

func TestStreaks_WeeklyBooleanPositive(t *testing.T) {
	h := &entities.Habit{
		Type:         value_objects.HabitTypeBoolean,
		Frequency:    value_objects.FrequencyWeekly,
		SpecificDays: []int{1, 3, 5},
		CreatedAt:    dt(2026, 3, 1),
	}

	t.Run("3 consecutive scheduled days", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 2), nil),
			makeEntry(dt(2026, 3, 4), nil),
			makeEntry(dt(2026, 3, 6), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 6))
		if r.Current != 3 || r.Longest != 3 {
			t.Errorf("Expected current=3 longest=3, got current=%d longest=%d", r.Current, r.Longest)
		}
	})

	t.Run("missed Wednesday breaks streak", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 2), nil),
			makeEntry(dt(2026, 3, 6), nil),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 6))
		if r.Current != 1 || r.Longest != 1 {
			t.Errorf("Expected current=1 longest=1, got current=%d longest=%d", r.Current, r.Longest)
		}
	})
}

func TestStreaks_CounterWithTarget(t *testing.T) {
	h := habit(value_objects.HabitTypeCounter, false, floatPtr(8))

	t.Run("all meet target", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 1), floatPtr(8)),
			makeEntry(dt(2026, 3, 2), floatPtr(10)),
			makeEntry(dt(2026, 3, 3), floatPtr(9)),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 3))
		if r.Current != 3 {
			t.Errorf("Expected current=3, got %d", r.Current)
		}
	})
}

func TestStreaks_WeeklyCounterNegativeWithTarget(t *testing.T) {
	h := &entities.Habit{
		Type:         value_objects.HabitTypeCounter,
		Frequency:    value_objects.FrequencyWeekly,
		SpecificDays: []int{1, 5},
		IsNegative:   true,
		TargetValue:  floatPtr(2),
		CreatedAt:    dt(2026, 3, 1),
	}

	t.Run("within limit and no entry both count as success", func(t *testing.T) {
		entries := []*entities.HabitEntry{
			makeEntry(dt(2026, 3, 2), floatPtr(1)),
			makeEntry(dt(2026, 3, 9), floatPtr(5)),
		}
		r := CalculateStreaks(entries, h, dt(2026, 3, 13))
		if r.Current != 1 || r.Longest != 2 {
			t.Errorf("Expected current=1 longest=2, got current=%d longest=%d", r.Current, r.Longest)
		}
	})
}
