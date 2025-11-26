package http

import (
	"encoding/json"
	"net/http"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/shared/errors"

	"github.com/go-chi/chi/v5"
)

type HabitHandlers struct {
	createHandler    *commands.CreateHabitHandler
	getTodaysHandler *queries.GetTodaysHabitsHandler
	markHandler      *commands.MarkHabitHandler
}

func NewHabitHandlers(
	createHandler *commands.CreateHabitHandler,
	getTodaysHandler *queries.GetTodaysHabitsHandler,
	markHandler *commands.MarkHabitHandler,
) *HabitHandlers {
	return &HabitHandlers{
		createHandler:    createHandler,
		getTodaysHandler: getTodaysHandler,
		markHandler:      markHandler,
	}
}

func (h *HabitHandlers) CreateHabit(w http.ResponseWriter, r *http.Request) {
	var req CreateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cmd := commands.CreateHabitCommand{
		UserID:        userID,
		Name:          req.Name,
		Description:   req.Description,
		Type:          req.Type,
		Frequency:     req.Frequency,
		SpecificDays:  req.SpecificDays,
		SpecificDates: req.SpecificDates,
		CarryOver:     req.CarryOver,
		TargetValue:   req.TargetValue,
	}

	habitID, err := h.createHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create habit")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"id": habitID})
}

func (h *HabitHandlers) GetTodaysHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	timezone := "UTC"

	query := queries.GetTodaysHabitsQuery{
		UserID:   userID,
		Timezone: timezone,
		Date:     time.Now().UTC(),
	}

	habits, err := h.getTodaysHandler.Handle(r.Context(), query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get habits")
		return
	}

	response := make([]TodaysHabitResponse, len(habits))
	for i, habit := range habits {
		response[i] = TodaysHabitResponse{
			ID:            habit.ID,
			Name:          habit.Name,
			Type:          habit.Type,
			TargetValue:   habit.TargetValue,
			ScheduledDate: habit.ScheduledDate,
			IsCarriedOver: habit.IsCarriedOver,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *HabitHandlers) MarkHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	var req MarkHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	scheduledDate, err := time.Parse("2006-01-02", req.ScheduledDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	cmd := commands.MarkHabitCommand{
		HabitID:       habitID,
		ScheduledDate: scheduledDate,
		Value:         req.Value,
	}

	if err := h.markHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrAlreadyExists {
			respondError(w, http.StatusConflict, "Habit already marked for this date")
			return
		}
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to mark habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "marked"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
