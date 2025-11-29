package http

import (
	"net/http"

	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/shared/errors"

	"apocapoc-api/internal/i18n"
	"github.com/go-chi/chi/v5"
)

type StatsHandlers struct {
	getHabitStatsHandler *queries.GetHabitStatsHandler
	translator           *i18n.Translator
}

func NewStatsHandlers(
	getHabitStatsHandler *queries.GetHabitStatsHandler,
	translator *i18n.Translator,
) *StatsHandlers {
	return &StatsHandlers{
		getHabitStatsHandler: getHabitStatsHandler,
		translator:           translator,
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
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	query := queries.GetHabitStatsQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	stats, err := h.getHabitStatsHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
