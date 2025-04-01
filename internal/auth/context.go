package auth

import "context"

// Context keys for storing authentication data in request context
type contextKey string

const (
	userIDContextKey   = contextKey("user_id")
	tenantIDContextKey = contextKey("tenant_id")
	userRoleContextKey = contextKey("user_role")
)

// GetUserIDContextKey returns the context key for user ID
func GetUserIDContextKey() contextKey {
	return userIDContextKey
}

// GetTenantIDContextKey returns the context key for tenant ID
func GetTenantIDContextKey() contextKey {
	return tenantIDContextKey
}

// GetUserRoleContextKey returns the context key for user role
func GetUserRoleContextKey() contextKey {
	return userRoleContextKey
}

// NewContextWithUserID adds user ID to context
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

// NewContextWithTenantID adds tenant ID to context
func NewContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDContextKey, tenantID)
}

// GetTenantIDFromContext retrieves tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantIDContextKey).(string)
	return tenantID, ok
}

// NewContextWithUserRole adds user role to context
func NewContextWithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleContextKey, role)
}

// GetUserRoleFromContext retrieves user role from context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleContextKey).(string)
	return role, ok
}
