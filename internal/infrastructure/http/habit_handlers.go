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
	createHandler          *commands.CreateHabitHandler
	getTodaysHandler       *queries.GetTodaysHabitsHandler
	getUserHabitsHandler   *queries.GetUserHabitsHandler
	getHabitByIDHandler    *queries.GetHabitByIDHandler
	getHabitEntriesHandler *queries.GetHabitEntriesHandler
	updateHandler          *commands.UpdateHabitHandler
	archiveHandler         *commands.ArchiveHabitHandler
	markHandler            *commands.MarkHabitHandler
	unmarkHandler          *commands.UnmarkHabitHandler
}

func NewHabitHandlers(
	createHandler *commands.CreateHabitHandler,
	getTodaysHandler *queries.GetTodaysHabitsHandler,
	getUserHabitsHandler *queries.GetUserHabitsHandler,
	getHabitByIDHandler *queries.GetHabitByIDHandler,
	getHabitEntriesHandler *queries.GetHabitEntriesHandler,
	updateHandler *commands.UpdateHabitHandler,
	archiveHandler *commands.ArchiveHabitHandler,
	markHandler *commands.MarkHabitHandler,
	unmarkHandler *commands.UnmarkHabitHandler,
) *HabitHandlers {
	return &HabitHandlers{
		createHandler:          createHandler,
		getTodaysHandler:       getTodaysHandler,
		getUserHabitsHandler:   getUserHabitsHandler,
		getHabitByIDHandler:    getHabitByIDHandler,
		getHabitEntriesHandler: getHabitEntriesHandler,
		updateHandler:          updateHandler,
		archiveHandler:         archiveHandler,
		markHandler:            markHandler,
		unmarkHandler:          unmarkHandler,
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

func (h *HabitHandlers) GetUserHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := queries.GetUserHabitsQuery{
		UserID: userID,
	}

	habits, err := h.getUserHabitsHandler.Handle(r.Context(), query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get habits")
		return
	}

	response := make([]UserHabitResponse, len(habits))
	for i, habit := range habits {
		response[i] = UserHabitResponse{
			ID:           habit.ID,
			Name:         habit.Name,
			Type:         habit.Type,
			Frequency:    habit.Frequency,
			SpecificDays: habit.SpecificDays,
			TargetValue:  habit.TargetValue,
			CarryOver:    habit.CarryOver,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *HabitHandlers) GetHabitByID(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := queries.GetHabitByIDQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	habit, err := h.getHabitByIDHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get habit")
		return
	}

	response := UserHabitResponse{
		ID:           habit.ID,
		Name:         habit.Name,
		Type:         habit.Type,
		Frequency:    habit.Frequency,
		SpecificDays: habit.SpecificDays,
		TargetValue:  habit.TargetValue,
		CarryOver:    habit.CarryOver,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *HabitHandlers) UpdateHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UpdateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := commands.UpdateHabitCommand{
		HabitID:       habitID,
		UserID:        userID,
		Name:          req.Name,
		Description:   req.Description,
		CarryOver:     req.CarryOver,
		TargetValue:   req.TargetValue,
		SpecificDays:  req.SpecificDays,
		SpecificDates: req.SpecificDates,
	}

	if err := h.updateHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		if err == errors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, "Invalid input")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HabitHandlers) ArchiveHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cmd := commands.ArchiveHabitCommand{
		HabitID: habitID,
		UserID:  userID,
	}

	if err := h.archiveHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to archive habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (h *HabitHandlers) GetHabitEntries(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query := queries.GetHabitEntriesQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	entries, err := h.getHabitEntriesHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get habit entries")
		return
	}

	response := make([]HabitEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = HabitEntryResponse{
			ID:            entry.ID,
			HabitID:       entry.HabitID,
			ScheduledDate: entry.ScheduledDate,
			CompletedAt:   entry.CompletedAt,
			Value:         entry.Value,
		}
	}

	respondJSON(w, http.StatusOK, response)
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

func (h *HabitHandlers) UnmarkHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")
	dateStr := chi.URLParam(r, "date")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	scheduledDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	cmd := commands.UnmarkHabitCommand{
		HabitID:       habitID,
		UserID:        userID,
		ScheduledDate: scheduledDate,
	}

	if err := h.unmarkHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrNotFound {
			respondError(w, http.StatusNotFound, "Habit entry not found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to unmark habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "unmarked"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
