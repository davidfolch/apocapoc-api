package http

import (
	"compress/gzip"
	"encoding/json"
	"net/http"

	"apocapoc-api/internal/application/queries"
	"apocapoc-api/internal/i18n"
)

type ExportHandlers struct {
	exportHandler *queries.ExportUserDataHandler
	translator    *i18n.Translator
}

func NewExportHandlers(
	exportHandler *queries.ExportUserDataHandler,
	translator *i18n.Translator,
) *ExportHandlers {
	return &ExportHandlers{
		exportHandler: exportHandler,
		translator:    translator,
	}
}

// ExportData godoc
// @Summary Export user data
// @Description Export all user habits and entries in JSON format with gzip compression. Limited to 1 export per hour.
// @Tags export
// @Security BearerAuth
// @Produce json
// @Success 200 {object} queries.ExportUserDataResult "Compressed JSON export"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 429 {object} ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /export [get]
func (h *ExportHandlers) ExportData(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	query := queries.ExportUserDataQuery{
		UserID: userID,
	}

	result, err := h.exportHandler.Handle(r.Context(), query)
	if err != nil {
		respondErrorI18n(w, r, h.translator, http.StatusInternalServerError, "export_failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"apocapoc-export.json.gz\"")
	w.WriteHeader(http.StatusOK)

	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	encoder := json.NewEncoder(gzipWriter)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return
	}
}
