package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	clients    map[string]*client
	rate       int // requests per second
	burst      int // maximum burst size
	mu         sync.Mutex
	cleanupInt time.Duration // cleanup interval
}

// client tracks rate limiting state for a single client
type client struct {
	tokens     int
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int, cleanupInterval time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		clients:    make(map[string]*client),
		rate:       rate,
		burst:      burst,
		cleanupInt: cleanupInterval,
	}

	// Start cleanup goroutine
	go limiter.cleanup(cleanupInterval)

	return limiter
}

// Allow returns true if the client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get or create client
	c, exists := rl.clients[clientIP]
	if !exists {
		c = &client{
			tokens:     rl.burst,
			lastUpdate: time.Now(),
		}
		rl.clients[clientIP] = c
	} else {
		// Add tokens based on time elapsed
		now := time.Now()
		elapsed := now.Sub(c.lastUpdate)
		c.lastUpdate = now

		// Calculate tokens to add based on elapsed time and rate
		newTokens := int(elapsed.Seconds() * float64(rl.rate))
		if newTokens > 0 {
			c.tokens += newTokens
			if c.tokens > rl.burst {
				c.tokens = rl.burst
			}
		}
	}

	// Check if client has tokens
	if c.tokens > 0 {
		c.tokens--
		return true
	}

	return false
}

// TimeUntilRefill returns seconds until the next token is available
func (rl *RateLimiter) TimeUntilRefill(clientIP string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[clientIP]
	if !exists || c.tokens > 0 {
		return 0 // No wait needed
	}

	// Calculate time needed for at least one token
	secondsNeeded := 1
	if rl.rate > 0 {
		secondsNeeded = 1 / rl.rate
		if secondsNeeded < 1 {
			secondsNeeded = 1
		}
	}

	return secondsNeeded
}

// cleanup removes clients that haven't been seen in a while
func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C

		rl.mu.Lock()
		deadline := time.Now().Add(-interval * 3) // Remove clients after 3 intervals
		for ip, client := range rl.clients {
			if client.lastUpdate.Before(deadline) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware that limits requests based on client IP
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr // Fallback if SplitHostPort fails
			}

			// Check if client is allowed
			if !limiter.Allow(ip) {
				// Calculate retry time
				retryAfter := limiter.TimeUntilRefill(ip)

				// Set headers
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded. Try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
