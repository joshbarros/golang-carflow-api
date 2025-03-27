package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler handles metrics requests
type Handler struct {
	metrics *Metrics
}

// NewHandler creates a new metrics handler
func NewHandler(metrics *Metrics) *Handler {
	return &Handler{
		metrics: metrics,
	}
}

// RegisterRoutes registers the metrics routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /metrics", h.GetMetrics)
}

// GetMetrics handles GET /metrics requests
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	stats := h.metrics.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// Middleware tracks metrics for each request
func Middleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Track the request
			metrics.IncrementRequestCount()

			// Create a custom response writer to capture the status code
			mrw := newMetricsResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(mrw, r)

			// Calculate request duration
			duration := time.Since(startTime)

			// Record response time
			metrics.AddResponseTime(duration)

			// Track errors
			if mrw.statusCode >= 400 {
				metrics.IncrementErrorCount()
			}

			// Record request info
			metrics.AddRequestInfo(RequestInfo{
				Path:      r.URL.Path,
				Method:    r.Method,
				Status:    mrw.statusCode,
				Duration:  duration,
				Timestamp: time.Now(),
			})
		})
	}
}

// metricsResponseWriter is a custom response writer that captures the status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newMetricsResponseWriter creates a new metrics response writer
func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

// WriteHeader captures the status code before writing it
func (mrw *metricsResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}
