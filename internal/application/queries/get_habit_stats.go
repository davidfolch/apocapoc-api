package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/shared/errors"
)

type HabitStatsDTO struct {
	HabitID            string  `json:"habit_id"`
	HabitName          string  `json:"habit_name"`
	TotalCompletions   int     `json:"total_completions"`
	CurrentStreak      int     `json:"current_streak"`
	LongestStreak      int     `json:"longest_streak"`
	CompletionRate     float64 `json:"completion_rate"`
	CompletionsThisWeek int    `json:"completions_this_week"`
	CompletionsThisMonth int   `json:"completions_this_month"`
}

type GetHabitStatsQuery struct {
	HabitID string
	UserID  string
}

type GetHabitStatsHandler struct {
	habitRepo repositories.HabitRepository
	entryRepo repositories.HabitEntryRepository
}

func NewGetHabitStatsHandler(
	habitRepo repositories.HabitRepository,
	entryRepo repositories.HabitEntryRepository,
) *GetHabitStatsHandler {
	return &GetHabitStatsHandler{
		habitRepo: habitRepo,
		entryRepo: entryRepo,
	}
}

func (h *GetHabitStatsHandler) Handle(ctx context.Context, query GetHabitStatsQuery) (*HabitStatsDTO, error) {
	habit, err := h.habitRepo.FindByID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	if habit.UserID != query.UserID {
		return nil, errors.ErrUnauthorized
	}

	entries, err := h.entryRepo.FindByHabitID(ctx, query.HabitID)
	if err != nil {
		return nil, err
	}

	stats := &HabitStatsDTO{
		HabitID:   habit.ID,
		HabitName: habit.Name,
	}

	if len(entries) == 0 {
		return stats, nil
	}

	stats.TotalCompletions = len(entries)
	stats.CurrentStreak = calculateCurrentStreak(entries)
	stats.LongestStreak = calculateLongestStreak(entries)
	stats.CompletionRate = calculateCompletionRate(entries, habit.CreatedAt)
	stats.CompletionsThisWeek = countCompletionsInPeriod(entries, 7)
	stats.CompletionsThisMonth = countCompletionsInPeriod(entries, 30)

	return stats, nil
}

func calculateCurrentStreak(entries []*entities.HabitEntry) int {
	if len(entries) == 0 {
		return 0
	}

	dateMap := make(map[string]bool)
	for _, entry := range entries {
		dateStr := entry.ScheduledDate.Format("2006-01-02")
		dateMap[dateStr] = true
	}

	streak := 0
	currentDate := time.Now().UTC()

	for {
		dateStr := currentDate.Format("2006-01-02")
		if !dateMap[dateStr] {
			break
		}
		streak++
		currentDate = currentDate.AddDate(0, 0, -1)
	}

	return streak
}

func calculateLongestStreak(entries []*entities.HabitEntry) int {
	if len(entries) == 0 {
		return 0
	}

	dateMap := make(map[string]bool)
	var dates []time.Time
	for _, entry := range entries {
		date := time.Date(entry.ScheduledDate.Year(), entry.ScheduledDate.Month(), entry.ScheduledDate.Day(), 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		if !dateMap[dateStr] {
			dateMap[dateStr] = true
			dates = append(dates, date)
		}
	}

	if len(dates) == 0 {
		return 0
	}

	longestStreak := 1
	currentStreak := 1

	for i := 1; i < len(dates); i++ {
		diff := dates[i].Sub(dates[i-1]).Hours() / 24
		if diff == 1 {
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}

	return longestStreak
}

func calculateCompletionRate(entries []*entities.HabitEntry, createdAt time.Time) float64 {
	if len(entries) == 0 {
		return 0
	}

	daysSinceCreation := int(time.Since(createdAt).Hours() / 24)
	if daysSinceCreation == 0 {
		daysSinceCreation = 1
	}

	rate := float64(len(entries)) / float64(daysSinceCreation) * 100
	if rate > 100 {
		rate = 100
	}

	return rate
}

func countCompletionsInPeriod(entries []*entities.HabitEntry, days int) int {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	count := 0

	for _, entry := range entries {
		if entry.ScheduledDate.After(cutoff) {
			count++
		}
	}

	return count
}
