package http

import (
	"encoding/json"
	"net/http"
	"time"

	"apocapoc-api/internal/application/commands"
	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/domain/entities"
	"apocapoc-api/internal/i18n"
	"apocapoc-api/internal/shared/errors"
)

type SyncHandlers struct {
	getSyncChangesHandler  *queries.GetSyncChangesHandler
	applySyncBatchHandler  *commands.ApplySyncBatchHandler
	translator             *i18n.Translator
}

func NewSyncHandlers(
	getSyncChangesHandler *queries.GetSyncChangesHandler,
	applySyncBatchHandler *commands.ApplySyncBatchHandler,
	translator *i18n.Translator,
) *SyncHandlers {
	return &SyncHandlers{
		getSyncChangesHandler: getSyncChangesHandler,
		applySyncBatchHandler: applySyncBatchHandler,
		translator:            translator,
	}
}

// GetSyncChanges godoc
// @Summary Get sync changes
// @Description Get all changes (habits and entries) since a given timestamp for offline sync
// @Tags sync
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param since query string true "ISO 8601 timestamp (e.g., 2025-01-01T00:00:00Z)"
// @Success 200 {object} SyncChangesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sync/changes [get]
func (h *SyncHandlers) GetSyncChanges(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "missing_since_parameter")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_since_format")
		return
	}

	query := queries.GetSyncChangesQuery{
		UserID: userID,
		Since:  since,
	}

	result, err := h.getSyncChangesHandler.Handle(r.Context(), query)
	if err != nil {
		if err == errors.ErrInvalidInput {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_input")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "internal_server_error")
		return
	}

	response := SyncChangesResponse{
		Habits: HabitChangesDTO{
			Created: toHabitDTOs(result.Habits.Created),
			Updated: toHabitDTOs(result.Habits.Updated),
			Deleted: result.Habits.Deleted,
		},
		Entries: EntryChangesDTO{
			Created: toHabitEntryDTOs(result.Entries.Created),
			Updated: toHabitEntryDTOs(result.Entries.Updated),
			Deleted: result.Entries.Deleted,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// ApplySyncBatch godoc
// @Summary Apply sync batch
// @Description Apply a batch of changes from the client for offline sync (Last-Write-Wins)
// @Tags sync
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SyncBatchRequest true "Sync batch data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sync/batch [post]
func (h *SyncHandlers) ApplySyncBatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondErrorI18n(w, r, h.translator, http.StatusUnauthorized, "user_not_authenticated")
		return
	}

	var req SyncBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_request_body")
		return
	}

	habitChanges := commands.HabitBatchChanges{
		Created: fromHabitDTOs(req.Habits.Created),
		Updated: fromHabitDTOs(req.Habits.Updated),
		Deleted: req.Habits.Deleted,
	}

	entryChanges := commands.EntryBatchChanges{
		Created: fromHabitEntryDTOs(req.Entries.Created),
		Updated: fromHabitEntryDTOs(req.Entries.Updated),
		Deleted: req.Entries.Deleted,
	}

	cmd := commands.ApplySyncBatchCommand{
		UserID:  userID,
		Habits:  habitChanges,
		Entries: entryChanges,
	}

	err := h.applySyncBatchHandler.Handle(r.Context(), cmd)
	if err != nil {
		if err == errors.ErrInvalidInput {
			respondErrorI18n(w, r, h.translator, http.StatusBadRequest, "invalid_input")
			return
		}
		if err == errors.ErrUnauthorized {
			respondErrorI18n(w, r, h.translator, http.StatusForbidden, "forbidden")
			return
		}
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "internal_server_error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "sync_batch_applied"})
}

func toHabitDTOs(habits []*entities.Habit) []SyncHabitDTO {
	dtos := make([]SyncHabitDTO, len(habits))
	for i, h := range habits {
		dtos[i] = SyncHabitDTO{
			ID:            h.ID,
			UserID:        h.UserID,
			Name:          h.Name,
			Description:   h.Description,
			Type:          h.Type,
			Frequency:     h.Frequency,
			SpecificDays:  h.SpecificDays,
			SpecificDates: h.SpecificDates,
			CarryOver:     h.CarryOver,
			IsNegative:    h.IsNegative,
			TargetValue:   h.TargetValue,
			CreatedAt:     h.CreatedAt,
			UpdatedAt:     h.UpdatedAt,
			ArchivedAt:    h.ArchivedAt,
		}
	}
	return dtos
}

func fromHabitDTOs(dtos []SyncHabitDTO) []*entities.Habit {
	habits := make([]*entities.Habit, len(dtos))
	for i, dto := range dtos {
		habits[i] = &entities.Habit{
			ID:            dto.ID,
			UserID:        dto.UserID,
			Name:          dto.Name,
			Description:   dto.Description,
			Type:          dto.Type,
			Frequency:     dto.Frequency,
			SpecificDays:  dto.SpecificDays,
			SpecificDates: dto.SpecificDates,
			CarryOver:     dto.CarryOver,
			IsNegative:    dto.IsNegative,
			TargetValue:   dto.TargetValue,
			CreatedAt:     dto.CreatedAt,
			UpdatedAt:     dto.UpdatedAt,
			ArchivedAt:    dto.ArchivedAt,
		}
	}
	return habits
}

func toHabitEntryDTOs(entries []*entities.HabitEntry) []SyncHabitEntryDTO {
	dtos := make([]SyncHabitEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = SyncHabitEntryDTO{
			ID:            e.ID,
			HabitID:       e.HabitID,
			ScheduledDate: e.ScheduledDate,
			CompletedAt:   e.CompletedAt,
			Value:         e.Value,
			UpdatedAt:     e.UpdatedAt,
		}
	}
	return dtos
}

func fromHabitEntryDTOs(dtos []SyncHabitEntryDTO) []*entities.HabitEntry {
	entries := make([]*entities.HabitEntry, len(dtos))
	for i, dto := range dtos {
		entries[i] = &entities.HabitEntry{
			ID:            dto.ID,
			HabitID:       dto.HabitID,
			ScheduledDate: dto.ScheduledDate,
			CompletedAt:   dto.CompletedAt,
			Value:         dto.Value,
			UpdatedAt:     dto.UpdatedAt,
		}
	}
	return entries
}
