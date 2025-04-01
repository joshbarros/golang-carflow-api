package auth

import (
	"net/http"
)

// AdminOnlyMiddleware restricts access to admin users
func AdminOnlyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user role from context
		role, ok := r.Context().Value(userRoleContextKey).(Role)
		if !ok || role != RoleAdmin {
			http.Error(w, "Unauthorized: Admin access required", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
