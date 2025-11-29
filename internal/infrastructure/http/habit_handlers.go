package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/domain/repositories"
	"apocapoc-api/internal/i18n"
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
	userRepo               repositories.UserRepository
	translator             *i18n.Translator
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
	userRepo repositories.UserRepository,
	translator *i18n.Translator,
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
		userRepo:               userRepo,
		translator:             translator,
	}
}

// CreateHabit godoc
// @Summary Create a new habit
// @Description Create a new habit for the authenticated user
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateHabitRequest true "Habit data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits [post]
func (h *HabitHandlers) CreateHabit(w http.ResponseWriter, r *http.Request) {
	var req CreateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_request_body")
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
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
		IsNegative:    req.IsNegative,
		TargetValue:   req.TargetValue,
	}

	habitID, err := h.createHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrInvalidInput {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_create_habit")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"id": habitID})
}

// GetUserHabits godoc
// @Summary Get all user habits
// @Description Get all active habits for the authenticated user
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserHabitResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits [get]
func (h *HabitHandlers) GetUserHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	query := queries.GetUserHabitsQuery{
		UserID: userID,
	}

	habits, err := h.getUserHabitsHandler.Handle(r.Context(), query)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_habits")
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
			IsNegative:   habit.IsNegative,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// GetHabitByID godoc
// @Summary Get habit by ID
// @Description Get a specific habit by ID
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Success 200 {object} UserHabitResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id} [get]
func (h *HabitHandlers) GetHabitByID(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	query := queries.GetHabitByIDQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	habit, err := h.getHabitByIDHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_habit")
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
		IsNegative:   habit.IsNegative,
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateHabit godoc
// @Summary Update habit
// @Description Update an existing habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Param request body UpdateHabitRequest true "Update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id} [put]
func (h *HabitHandlers) UpdateHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	var req UpdateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_request_body")
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
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		if err == errors.ErrInvalidInput {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_input")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_update_habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ArchiveHabit godoc
// @Summary Archive habit
// @Description Archive (soft delete) a habit
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id} [delete]
func (h *HabitHandlers) ArchiveHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	cmd := commands.ArchiveHabitCommand{
		HabitID: habitID,
		UserID:  userID,
	}

	if err := h.archiveHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_archive_habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

// GetHabitEntries godoc
// @Summary Get habit entries
// @Description Get entries (completion history) for a habit with optional date filtering and pagination
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param limit query int false "Page size (max 100)"
// @Success 200 {object} HabitEntriesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id}/entries [get]
func (h *HabitHandlers) GetHabitEntries(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	query := queries.GetHabitEntriesQuery{
		HabitID: habitID,
		UserID:  userID,
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_from_date_format")
			return
		}
		query.From = &from
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_to_date_format")
			return
		}
		query.To = &to
	}

	var dateRangeDays int
	if query.From != nil && query.To != nil {
		dateRangeDays = int(query.To.Sub(*query.From).Hours() / 24)
	}

	requiresPagination := false
	if query.From == nil || query.To == nil {
		requiresPagination = true
	} else if dateRangeDays > 365 {
		requiresPagination = true
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_page_parameter")
			return
		}
		query.Page = page
	} else if requiresPagination {
		query.Page = 1
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_limit_parameter")
			return
		}
		query.Limit = limit
	} else if requiresPagination {
		query.Limit = 50
	}

	if requiresPagination && query.Limit == 0 {
		respondError(w, http.StatusBadRequest, "Pagination required: provide 'limit' parameter or use date range \u2264 1 year")
		return
	}

	result, err := h.getHabitEntriesHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_habit_entries")
		return
	}

	entries := make([]HabitEntryResponse, len(result.Entries))
	for i, entry := range result.Entries {
		entries[i] = HabitEntryResponse{
			ID:            entry.ID,
			HabitID:       entry.HabitID,
			ScheduledDate: entry.ScheduledDate,
			CompletedAt:   entry.CompletedAt,
			Value:         entry.Value,
		}
	}

	response := HabitEntriesResponse{
		Entries: entries,
		Total:   result.Total,
		Page:    result.Page,
		Limit:   result.Limit,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetTodaysHabits godoc
// @Summary Get today's habits
// @Description Get all habits scheduled for today for the authenticated user. Includes the entry for today if it exists.
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Success 200 {array} TodaysHabitResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/today [get]
func (h *HabitHandlers) GetTodaysHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_user")
		return
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}

	today := time.Now().In(loc)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	query := queries.GetTodaysHabitsQuery{
		UserID:   userID,
		Timezone: user.Timezone,
		Date:     todayDate,
	}

	habits, err := h.getTodaysHandler.Handle(r.Context(), query)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_get_habits")
		return
	}

	response := make([]TodaysHabitResponse, len(habits))
	for i, habit := range habits {
		var entryResponse *TodaysHabitEntryResponse
		if habit.Entry != nil {
			entryResponse = &TodaysHabitEntryResponse{
				ID:          habit.Entry.ID,
				Value:       habit.Entry.Value,
				CompletedAt: habit.Entry.CompletedAt,
			}
		}

		response[i] = TodaysHabitResponse{
			ID:            habit.ID,
			Name:          habit.Name,
			Type:          habit.Type,
			TargetValue:   habit.TargetValue,
			IsNegative:    habit.IsNegative,
			ScheduledDate: habit.ScheduledDate,
			IsCarriedOver: habit.IsCarriedOver,
			Entry:         entryResponse,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// MarkHabit godoc
// @Summary Mark habit as complete
// @Description Mark a habit as completed for a specific date
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Param request body MarkHabitRequest true "Mark data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id}/mark [post]
func (h *HabitHandlers) MarkHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")

	var req MarkHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_request_body")
		return
	}

	scheduledDate, err := time.Parse("2006-01-02", req.ScheduledDate)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_date_format")
		return
	}

	cmd := commands.MarkHabitCommand{
		HabitID:       habitID,
		ScheduledDate: scheduledDate,
		Value:         req.Value,
	}

	if err := h.markHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrAlreadyExists {
			respondErrorI18n(w, r, h.translator, http.StatusConflict, "habit_already_marked")
			return
		}
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_not_found")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_mark_habit")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "marked"})
}

// UnmarkHabit godoc
// @Summary Unmark habit
// @Description Delete a habit entry (unmark completion)
// @Tags habits
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID"
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /habits/{id}/entries/{date} [delete]
func (h *HabitHandlers) UnmarkHabit(w http.ResponseWriter, r *http.Request) {
	habitID := chi.URLParam(r, "id")
	dateStr := chi.URLParam(r, "date")

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	scheduledDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_date_format")
		return
	}

	cmd := commands.UnmarkHabitCommand{
		HabitID:       habitID,
		UserID:        userID,
		ScheduledDate: scheduledDate,
	}

	if err := h.unmarkHandler.Handle(r.Context(), cmd); err != nil {
		if err == errors.ErrNotFound {
			respondErrorI18n(w, r, h.translator, http.StatusNotFound, "habit_entry_not_found")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "access_denied")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "failed_unmark_habit")
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

func respondValidationError(w http.ResponseWriter, err error) {
	errMsg := err.Error()
	var field string

	if strings.Contains(errMsg, ": ") {
		parts := strings.SplitN(errMsg, ": ", 3)
		if len(parts) >= 3 {
			field = parts[1]
			errMsg = parts[2]
			respondJSON(w, http.StatusBadRequest, ValidationErrorResponse{
				Error: errMsg,
				Field: field,
			})
			return
		}
	}

	respondJSON(w, http.StatusBadRequest, ErrorResponse{
		Error: errMsg,
	})
}
