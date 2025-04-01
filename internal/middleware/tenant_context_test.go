package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockDB is a mock implementation of the database interface
type mockDB struct {
	execContext func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (db *mockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.execContext(ctx, query, args...)
}

func TestTenantContextMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		tenantID   string
		setupDB    func() *mockDB
		wantStatus int
	}{
		{
			name:     "Valid tenant ID",
			tenantID: "test-tenant",
			setupDB: func() *mockDB {
				return &mockDB{
					execContext: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
						return nil, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "Missing tenant ID",
			tenantID: "",
			setupDB: func() *mockDB {
				return &mockDB{
					execContext: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
						return nil, nil
					},
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "Database error",
			tenantID: "test-tenant",
			setupDB: func() *mockDB {
				return &mockDB{
					execContext: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
						return nil, sql.ErrConnDone
					},
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if tenant ID is set in context
				tenantID := r.Context().Value(TenantIDContextKey)
				if tenantID != tt.tenantID && tt.wantStatus == http.StatusOK {
					t.Errorf("TenantID in context = %v, want %v", tenantID, tt.tenantID)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			db := tt.setupDB()
			middleware := NewTenantContextMiddleware(db)

			// Create request with tenant ID in header
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.tenantID != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantID)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply middleware
			handler := middleware(nextHandler)
			handler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestTenantContextMiddlewareCleanup(t *testing.T) {
	// Create test handler that simulates a panic
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create mock DB that tracks SET LOCAL calls
	var setLocalCalled, resetLocalCalled bool
	db := &mockDB{
		execContext: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			if query == "SET LOCAL tenant.id = $1" {
				setLocalCalled = true
			} else if query == "RESET LOCAL tenant.id" {
				resetLocalCalled = true
			}
			return nil, nil
		},
	}

	// Create middleware
	middleware := NewTenantContextMiddleware(db)

	// Create request with tenant ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply middleware and expect panic to be recovered
	handler := middleware(nextHandler)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic was not recovered")
			}
		}()
		handler.ServeHTTP(w, req)
	}()

	// Check if both SET LOCAL and RESET LOCAL were called
	if !setLocalCalled {
		t.Error("SET LOCAL was not called")
	}
	if !resetLocalCalled {
		t.Error("RESET LOCAL was not called")
	}
}

func TestTenantContextMiddlewareWithTransaction(t *testing.T) {
	// Create test handler that checks tenant ID in transaction
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real application, this would be a transaction
		w.WriteHeader(http.StatusOK)
	})

	// Create mock DB that tracks queries in transaction
	var queries []string
	db := &mockDB{
		execContext: func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
			queries = append(queries, query)
			return nil, nil
		},
	}

	// Create middleware
	middleware := NewTenantContextMiddleware(db)

	// Create request with tenant ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply middleware
	handler := middleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Check if queries were executed in correct order
	expectedQueries := []string{
		"SET LOCAL tenant.id = $1",
		"RESET LOCAL tenant.id",
	}

	if len(queries) != len(expectedQueries) {
		t.Errorf("Got %d queries, want %d", len(queries), len(expectedQueries))
	}

	for i, query := range queries {
		if i < len(expectedQueries) && query != expectedQueries[i] {
			t.Errorf("Query %d = %q, want %q", i, query, expectedQueries[i])
		}
	}
}
