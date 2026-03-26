package services

import (
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/shared/utils"
)

type StreakResult struct {
	Current int
	Longest int
}

func CalculateStreaks(entries []*entities.HabitEntry, habit *entities.Habit, now time.Time) StreakResult {
	entryMap := buildEntryMap(entries)
	scheduled := allScheduledDates(habit, now)

	if len(scheduled) == 0 {
		return StreakResult{}
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	longest := 0
	current := 0

	for _, d := range scheduled {
		if IsDaySuccessful(d.Format("2006-01-02"), entryMap, habit) {
			current++
			if current > longest {
				longest = current
			}
		} else if d.Equal(today) && !habit.IsNegative {
			continue
		} else {
			current = 0
		}
	}

	return StreakResult{Current: current, Longest: longest}
}

func buildEntryMap(entries []*entities.HabitEntry) map[string]*entities.HabitEntry {
	m := make(map[string]*entities.HabitEntry)
	for _, e := range entries {
		m[e.ScheduledDate.Format("2006-01-02")] = e
	}
	return m
}

func allScheduledDates(habit *entities.Habit, now time.Time) []time.Time {
	createdUTC := habit.CreatedAt.UTC()
	start := time.Date(createdUTC.Year(), createdUTC.Month(), createdUTC.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	freq := string(habit.Frequency)

	var dates []time.Time
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		if utils.ShouldAppearToday(freq, habit.SpecificDays, habit.SpecificDates, d) {
			dates = append(dates, d)
		}
	}
	return dates
}

func IsDaySuccessful(dateStr string, entryMap map[string]*entities.HabitEntry, habit *entities.Habit) bool {
	entry, hasEntry := entryMap[dateStr]
	habitType := string(habit.Type)

	if !habit.IsNegative {
		if !hasEntry {
			return false
		}
		if habit.TargetValue != nil && entry.Value != nil {
			return *entry.Value >= *habit.TargetValue
		}
		return true
	}

	if habitType == "VALUE" && habit.TargetValue != nil {
		if !hasEntry {
			return false
		}
		if entry.Value != nil {
			return *entry.Value <= *habit.TargetValue
		}
		return false
	}

	if !hasEntry {
		return true
	}

	if habit.TargetValue != nil && entry.Value != nil {
		return *entry.Value <= *habit.TargetValue
	}

	return false
}
