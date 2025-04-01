package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminOnlyMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		wantStatus int
	}{
		{
			name:       "Admin user",
			role:       RoleAdmin,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Non-admin user",
			role:       RoleUser,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Missing role",
			role:       "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request with role in context
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), UserRoleContextKey, tt.role)
			req = req.WithContext(ctx)

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply middleware
			handler := AdminOnlyMiddleware(nextHandler)
			handler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAdminOnlyMiddlewareWithMissingContext(t *testing.T) {
	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request without role in context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Apply middleware
	handler := AdminOnlyMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusForbidden {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestAdminOnlyMiddlewareWithInvalidContextValue(t *testing.T) {
	// Create test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request with invalid role type in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserRoleContextKey, 123) // Invalid type
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Apply middleware
	handler := AdminOnlyMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusForbidden {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestAdminOnlyMiddlewareWithPanic(t *testing.T) {
	// Create test handler that panics
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create request with admin role
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserRoleContextKey, RoleAdmin)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Apply middleware and expect panic to be propagated
	handler := AdminOnlyMiddleware(nextHandler)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic was not propagated")
		}
	}()
	handler.ServeHTTP(w, req)
}
