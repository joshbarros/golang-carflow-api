package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/joshbarros/golang-carflow-api/internal/interfaces"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting per tenant
type RateLimiter struct {
	limiters  map[string]*rate.Limiter
	mu        sync.RWMutex
	tenantSvc interfaces.TenantService
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(tenantSvc interfaces.TenantService) *RateLimiter {
	return &RateLimiter{
		limiters:  make(map[string]*rate.Limiter),
		tenantSvc: tenantSvc,
	}
}

// getLimiter returns a rate limiter for a tenant
func (rl *RateLimiter) getLimiter(tenantID string) (*rate.Limiter, error) {
	rl.mu.RLock()
	limiter, exists := rl.limiters[tenantID]
	rl.mu.RUnlock()

	if exists {
		return limiter, nil
	}

	// Get tenant configuration
	t, err := rl.tenantSvc.GetTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %v", err)
	}

	// Create new limiter based on tenant's plan
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check again in case another goroutine created it
	if limiter, exists = rl.limiters[tenantID]; exists {
		return limiter, nil
	}

	// Convert requests per minute to requests per second
	rps := float64(t.Limits.APIRateLimit) / 60.0
	limiter = rate.NewLimiter(rate.Limit(rps), int(rps)) // Burst size equals one second worth of requests
	rl.limiters[tenantID] = limiter

	return limiter, nil
}

// Middleware creates a middleware that enforces rate limits
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant ID from context
		tenantID, ok := r.Context().Value(TenantIDContextKey).(string)
		if !ok {
			http.Error(w, "Unauthorized: Tenant ID not found", http.StatusUnauthorized)
			return
		}

		// Get limiter for tenant
		limiter, err := rl.getLimiter(tenantID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if request is allowed
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// cleanupLimiters periodically removes unused limiters
func (rl *RateLimiter) cleanupLimiters(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			// In a production environment, you might want to track last usage time
			// and only remove limiters that haven't been used for a while
			rl.limiters = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}
}

// RateLimitMiddleware creates a middleware that limits requests based on tenant
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant ID from context
			tenantID, ok := r.Context().Value(TenantIDContextKey).(string)
			if !ok {
				http.Error(w, "Unauthorized: Tenant ID not found", http.StatusUnauthorized)
				return
			}

			// Get limiter for tenant
			tenantLimiter, err := limiter.getLimiter(tenantID)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Check if request is allowed
			if !tenantLimiter.Allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
