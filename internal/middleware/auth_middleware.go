package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/joshbarros/golang-carflow-api/internal/auth"
)

// AuthContextKeys for storing authentication data in request context
type AuthContextKey string

const (
	UserIDContextKey   AuthContextKey = "user_id"
	TenantIDContextKey AuthContextKey = "tenant_id"
	UserRoleContextKey AuthContextKey = "user_role"
)

// AuthMiddleware is a middleware that validates JWT tokens
type AuthMiddleware struct {
	tokenService *auth.TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: auth.NewTokenService(),
	}
}

// Middleware creates a middleware that requires authentication
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from bearer prefix
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid authorization format", http.StatusUnauthorized)
			return
		}
		tokenString := tokenParts[1]

		// Validate token
		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Create a new request with the user info in the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantIDContextKey, claims.TenantID)
		ctx = context.WithValue(ctx, UserRoleContextKey, string(claims.Role))

		// Call the next handler with the enhanced context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnlyMiddleware creates a middleware that requires admin role
func (m *AuthMiddleware) AdminOnlyMiddleware(next http.Handler) http.Handler {
	return m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user role from context
		role, ok := r.Context().Value(UserRoleContextKey).(string)
		if !ok {
			http.Error(w, "Unauthorized: Role not found in token", http.StatusUnauthorized)
			return
		}

		// Ensure user is an admin
		if role != string(auth.RoleAdmin) {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}))
}

// TenantScopeMiddleware creates a middleware that ensures the user belongs to the right tenant
func (m *AuthMiddleware) TenantScopeMiddleware(next http.Handler) http.Handler {
	return m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant ID from context
		tenantID, ok := r.Context().Value(TenantIDContextKey).(string)
		if !ok {
			http.Error(w, "Unauthorized: Tenant ID not found in token", http.StatusUnauthorized)
			return
		}

		// Get tenant ID from request (in a real app, this might come from a path parameter)
		requestTenantID := r.URL.Query().Get("tenant_id")
		if requestTenantID == "" {
			requestTenantID = "default" // Default tenant
		}

		// Ensure tenant IDs match or user is admin
		role, _ := r.Context().Value(UserRoleContextKey).(string)
		isAdmin := role == string(auth.RoleAdmin)

		if tenantID != requestTenantID && !isAdmin {
			http.Error(w, "Forbidden: Access denied for this tenant", http.StatusForbidden)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}))
}
