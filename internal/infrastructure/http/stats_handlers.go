package http

import (
	"net/http"

	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/shared/errors"

	"github.com/go-chi/chi/v5"
)

type StatsHandlers struct {
	getHabitStatsHandler *queries.GetHabitStatsHandler
}

func NewStatsHandlers(
	getHabitStatsHandler *queries.GetHabitStatsHandler,
) *StatsHandlers {
	return &StatsHandlers{
		getHabitStatsHandler: getHabitStatsHandler,
	}
}

// GetHabitStats godoc
// @Summary Get habit statistics
// @Description Get statistics for a specific habit including streaks and completion rates
// @Tags stats
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Success 200 {object} queries.HabitStatsDTO
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/habits/{id} [get]
func (h *StatsHandlers) GetHabitStats(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := queries.GetHabitStatsQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	stats, err := h.getHabitStatsHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get habit stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
