package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs information about each HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create a custom response writer to capture the status code
		lrw := newLoggingResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Calculate request duration
		duration := time.Since(startTime)

		// Log the request details
		log.Printf(
			"[%s] %s %s %d %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			lrw.statusCode,
			duration,
		)
	})
}

// loggingResponseWriter is a custom response writer that captures the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new logging response writer
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

// WriteHeader captures the status code before writing it
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
