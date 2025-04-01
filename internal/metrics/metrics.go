package metrics

import (
	"net/http"
	"sync"
	"time"
)

// Metrics tracks application metrics
type Metrics struct {
	RequestCount  int64
	ErrorCount    int64
	ResponseTimes []time.Duration
	LastRequests  []RequestInfo
	StartTime     time.Time
	mu            sync.RWMutex
}

// RequestInfo contains information about a request
type RequestInfo struct {
	Path      string
	Method    string
	Status    int
	Duration  time.Duration
	Timestamp time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		ResponseTimes: make([]time.Duration, 0, 100),
		LastRequests:  make([]RequestInfo, 0, 10),
		StartTime:     time.Now(),
	}
}

// IncrementRequestCount increments the request counter
func (m *Metrics) IncrementRequestCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RequestCount++
}

// IncrementErrorCount increments the error counter
func (m *Metrics) IncrementErrorCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCount++
}

// AddResponseTime adds a response time measurement
func (m *Metrics) AddResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Keep only the last 100 response times for percentile calculations
	if len(m.ResponseTimes) >= 100 {
		m.ResponseTimes = m.ResponseTimes[1:]
	}
	m.ResponseTimes = append(m.ResponseTimes, duration)
}

// AddRequestInfo adds information about a request
func (m *Metrics) AddRequestInfo(info RequestInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Keep only the last 10 requests
	if len(m.LastRequests) >= 10 {
		m.LastRequests = m.LastRequests[1:]
	}
	m.LastRequests = append(m.LastRequests, info)
}

// GetStats gets the current metrics
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"requests": map[string]interface{}{
			"total":  m.RequestCount,
			"errors": m.ErrorCount,
		},
		"uptime":        time.Since(m.StartTime).String(),
		"last_requests": m.LastRequests,
	}

	// Calculate response time percentiles if we have enough data
	if len(m.ResponseTimes) > 0 {
		// Make a copy to avoid modifying the original
		times := make([]time.Duration, len(m.ResponseTimes))
		copy(times, m.ResponseTimes)

		// Sort the times
		timeStats := calculateTimeStats(times)
		stats["response_times"] = timeStats
	}

	return stats
}

// calculateTimeStats calculates statistics for response times
func calculateTimeStats(times []time.Duration) map[string]interface{} {
	var total time.Duration
	for _, t := range times {
		total += t
	}

	avg := total / time.Duration(len(times))

	return map[string]interface{}{
		"avg":   avg.String(),
		"count": len(times),
	}
}

// MetricsMiddleware creates a middleware that tracks request metrics
func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture the status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start)
			metrics.IncrementRequestCount()
			if rw.statusCode >= 400 {
				metrics.IncrementErrorCount()
			}
			metrics.AddResponseTime(duration)
			metrics.AddRequestInfo(RequestInfo{
				Path:      r.URL.Path,
				Method:    r.Method,
				Status:    rw.statusCode,
				Duration:  duration,
				Timestamp: start,
			})
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
