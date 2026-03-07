package queries

import (
	"context"
	"time"

	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/domain/services"
	"apocapoc-api/internal/shared/errors"
)

type HabitStatsDTO struct {
	HabitID              string `json:"habit_id"`
	HabitName            string `json:"habit_name"`
	TotalCompletions     int    `json:"total_completions"`
	CurrentStreak        int    `json:"current_streak"`
	LongestStreak        int    `json:"longest_streak"`
	CompletionsThisWeek  int    `json:"completions_this_week"`
	CompletionsThisMonth int    `json:"completions_this_month"`
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

	if len(entries) == 0 && !habit.IsNegative {
		return stats, nil
	}

	stats.TotalCompletions = len(entries)
	stats.CompletionsThisWeek = countCompletionsInPeriod(entries, 7)
	stats.CompletionsThisMonth = countCompletionsInPeriod(entries, 30)

	streaks := services.CalculateStreaks(entries, habit, time.Now().UTC())
	stats.CurrentStreak = streaks.Current
	stats.LongestStreak = streaks.Longest

	return stats, nil
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
