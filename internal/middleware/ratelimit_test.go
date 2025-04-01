package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

// mockTenantService is a mock implementation of the tenant service interface
type mockTenantService struct {
	tenant *domain.Tenant
}

func (s *mockTenantService) CreateTenant(tenant domain.Tenant) error {
	return nil
}

func (s *mockTenantService) GetTenant(id string) (*domain.Tenant, error) {
	if s.tenant != nil && s.tenant.ID == id {
		return s.tenant, nil
	}
	return nil, domain.ErrTenantNotFound
}

func (s *mockTenantService) UpdateTenant(tenant domain.Tenant) error {
	return nil
}

func (s *mockTenantService) DeleteTenant(id string) error {
	return nil
}

func (s *mockTenantService) ListTenants(page, pageSize int) ([]domain.Tenant, error) {
	return nil, nil
}

func (s *mockTenantService) GetTenantByDomain(domain string) (*domain.Tenant, error) {
	return nil, nil
}

func TestRateLimiter(t *testing.T) {
	// Create a mock tenant service
	mockService := &mockTenantService{
		tenant: &domain.Tenant{
			ID:   uuid.New().String(),
			Plan: "basic",
			Limits: domain.ResourceLimits{
				APIRateLimit: 120, // 2 requests per second
			},
		},
	}

	// Create a rate limiter
	limiter := NewRateLimiter(mockService)

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name          string
		tenantID      string
		requestCount  int
		sleepBetween  time.Duration
		wantStatus    []int
		cleanupBefore bool
	}{
		{
			name:         "Within rate limit",
			tenantID:     mockService.tenant.ID,
			requestCount: 2,
			sleepBetween: 0,
			wantStatus:   []int{http.StatusOK, http.StatusOK},
		},
		{
			name:         "Exceeds rate limit",
			tenantID:     mockService.tenant.ID,
			requestCount: 3,
			sleepBetween: 0,
			wantStatus:   []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
		{
			name:         "Rate limit resets after wait",
			tenantID:     mockService.tenant.ID,
			requestCount: 3,
			sleepBetween: time.Second,
			wantStatus:   []int{http.StatusOK, http.StatusOK, http.StatusOK},
		},
		{
			name:         "Invalid tenant ID",
			tenantID:     "invalid",
			requestCount: 1,
			sleepBetween: 0,
			wantStatus:   []int{http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanupBefore {
				ctx := context.Background()
				limiter.cleanupLimiters(ctx)
			}

			for i := 0; i < tt.requestCount; i++ {
				// Create request with tenant ID in context
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(req.Context(), TenantIDContextKey, tt.tenantID)
				req = req.WithContext(ctx)

				// Create response recorder
				w := httptest.NewRecorder()

				// Apply rate limiter middleware
				handler := limiter.Middleware(nextHandler)
				handler.ServeHTTP(w, req)

				// Check response status
				if w.Code != tt.wantStatus[i] {
					t.Errorf("Request %d: got status %d, want %d", i+1, w.Code, tt.wantStatus[i])
				}

				// Sleep between requests if specified
				if tt.sleepBetween > 0 {
					time.Sleep(tt.sleepBetween)
				}
			}
		})
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	// Create a mock tenant service
	mockService := &mockTenantService{
		tenant: &domain.Tenant{
			ID:   uuid.New().String(),
			Plan: "basic",
			Limits: domain.ResourceLimits{
				APIRateLimit: 120, // 2 requests per second
			},
		},
	}

	// Create a rate limiter
	limiter := NewRateLimiter(mockService)

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Make a request to create a limiter
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), TenantIDContextKey, mockService.tenant.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler := limiter.Middleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Run cleanup
	cleanupCtx := context.Background()
	limiter.cleanupLimiters(cleanupCtx)

	// Check if limiter was cleaned up
	limiter.mu.Lock()
	_, exists := limiter.limiters[mockService.tenant.ID]
	limiter.mu.Unlock()

	if exists {
		t.Error("Limiter was not cleaned up")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	// Create a mock tenant service
	mockService := &mockTenantService{
		tenant: &domain.Tenant{
			ID:   uuid.New().String(),
			Plan: "basic",
			Limits: domain.ResourceLimits{
				APIRateLimit: 600, // 10 requests per second
			},
		},
	}

	// Create a rate limiter
	limiter := NewRateLimiter(mockService)

	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create multiple concurrent requests
	concurrentRequests := 20
	done := make(chan bool)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), TenantIDContextKey, mockService.tenant.ID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler := limiter.Middleware(nextHandler)
			handler.ServeHTTP(w, req)

			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}

	// Check if limiter still exists and is working
	limiter.mu.Lock()
	l, exists := limiter.limiters[mockService.tenant.ID]
	limiter.mu.Unlock()

	if !exists {
		t.Error("Limiter was unexpectedly removed")
	}

	if l == nil {
		t.Error("Limiter is nil")
	}
}
