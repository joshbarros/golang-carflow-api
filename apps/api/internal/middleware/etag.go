package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net/http"
)

// ETagWriter is a custom response writer that captures the response for ETag generation
type ETagWriter struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

// NewETagWriter creates a new ETag writer
func NewETagWriter(w http.ResponseWriter) *ETagWriter {
	return &ETagWriter{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
		status:         http.StatusOK, // Default status code
	}
}

// WriteHeader captures the status code
func (e *ETagWriter) WriteHeader(code int) {
	e.status = code
	// Don't write header yet, it will be written when we flush
}

// Write captures the response body
func (e *ETagWriter) Write(b []byte) (int, error) {
	return e.buf.Write(b)
}

// generateETag generates an ETag from the response body
func (e *ETagWriter) generateETag() string {
	hash := md5.Sum(e.buf.Bytes())
	return hex.EncodeToString(hash[:])
}

// Flush writes the actual response with ETag header
func (e *ETagWriter) Flush(r *http.Request) {
	// Only add ETag for successful GET responses
	if e.status == http.StatusOK {
		etag := e.generateETag()

		// Check If-None-Match header
		if match := r.Header.Get("If-None-Match"); match == etag {
			e.ResponseWriter.WriteHeader(http.StatusNotModified)
			return
		}

		// Set ETag header
		e.ResponseWriter.Header().Set("ETag", etag)
	}

	// Write status code and headers
	e.ResponseWriter.WriteHeader(e.status)

	// Write the body
	e.buf.WriteTo(e.ResponseWriter)
}

// ETagMiddleware adds ETag support for caching
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only add ETag for GET requests
		if r.Method == http.MethodGet {
			etw := NewETagWriter(w)
			next.ServeHTTP(etw, r)
			etw.Flush(r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
