package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler is a health check handler
type Handler struct {
	startTime time.Time
}

// NewHandler creates a new health check handler
func NewHandler() *Handler {
	return &Handler{
		startTime: time.Now(),
	}
}

// RegisterRoutes registers the health check routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.HealthCheck)
}

// HealthCheck handles GET /healthz requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"uptime":    time.Since(h.startTime).String(),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
