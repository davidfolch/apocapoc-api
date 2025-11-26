package http

import (
	"database/sql"
	"net/http"
	"time"
)

var startTime = time.Now()

type HealthHandlers struct {
	db *sql.DB
}

func NewHealthHandlers(db *sql.DB) *HealthHandlers {
	return &HealthHandlers{
		db: db,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Uptime   string `json:"uptime"`
}

// Health godoc
// @Summary Health check
// @Description Get API health status including database connectivity and uptime
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandlers) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	overallStatus := "ok"
	statusCode := http.StatusOK

	if err := h.db.Ping(); err != nil {
		dbStatus = "error"
		overallStatus = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	uptime := time.Since(startTime)
	uptimeStr := formatDuration(uptime)

	response := HealthResponse{
		Status:   overallStatus,
		Database: dbStatus,
		Uptime:   uptimeStr,
	}

	respondJSON(w, statusCode, response)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return formatTime(int(h), "h") + formatTime(int(m), "m")
	}
	if m > 0 {
		return formatTime(int(m), "m") + formatTime(int(s), "s")
	}
	return formatTime(int(s), "s")
}

func formatTime(value int, unit string) string {
	if value == 0 {
		return ""
	}
	return formatInt(value) + unit
}

func formatInt(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
