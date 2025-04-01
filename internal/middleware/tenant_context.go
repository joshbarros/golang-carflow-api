package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
)

// TenantContextMiddleware sets the tenant context for database queries
type TenantContextMiddleware struct {
	db *sql.DB
}

// NewTenantContextMiddleware creates a new tenant context middleware
func NewTenantContextMiddleware(db *sql.DB) *TenantContextMiddleware {
	return &TenantContextMiddleware{
		db: db,
	}
}

// Middleware creates a middleware that sets the tenant context
func (m *TenantContextMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant ID from context (set by auth middleware)
		tenantID, ok := r.Context().Value(TenantIDContextKey).(string)
		if !ok {
			http.Error(w, "Unauthorized: Tenant ID not found", http.StatusUnauthorized)
			return
		}

		// Set tenant ID in database session
		if _, err := m.db.ExecContext(r.Context(), "SET app.tenant_id = $1", tenantID); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create a cleanup function to reset the tenant ID
		cleanup := func() {
			if _, err := m.db.ExecContext(context.Background(), "RESET app.tenant_id"); err != nil {
				fmt.Printf("Error resetting tenant ID: %v\n", err)
			}
		}

		// Ensure cleanup runs after the request
		defer cleanup()

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
