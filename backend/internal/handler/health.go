package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yuryalencar/research-events/internal/health"
)

// --- Types ---

// HealthHandler serves GET /health.
// Dependencies are injected via constructor — never use globals.
type HealthHandler struct {
	registry   *health.Registry
	startTime  time.Time
	appVersion string
}

// --- Constructor ---

func NewHealthHandler(registry *health.Registry, appVersion string) *HealthHandler {
	return &HealthHandler{
		registry:   registry,
		startTime:  time.Now(),
		appVersion: appVersion,
	}
}

// --- Public methods ---

// ServeHTTP handles GET /health.
// Returns 200 when all checkers pass, 503 when any checker fails.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	checks, allHealthy := h.registry.RunAll(r.Context())

	status := health.StatusHealthy
	httpStatus := http.StatusOK
	if !allHealthy {
		status = health.StatusUnhealthy
		httpStatus = http.StatusServiceUnavailable
	}

	body := health.HealthResponse{
		Status:    status,
		Version:   h.appVersion,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Checks:    checks,
	}

	h.writeJSON(w, httpStatus, body)
}

// --- Private helpers ---

func (h *HealthHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
