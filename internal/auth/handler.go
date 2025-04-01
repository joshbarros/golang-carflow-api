package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// ResponseError represents an error response
type ResponseError struct {
	Error string `json:"error"`
}

// Handler handles auth-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers auth routes to the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.HandleRegister)
	mux.HandleFunc("POST /auth/login", h.HandleLogin)
	mux.HandleFunc("POST /auth/refresh", h.HandleRefreshToken)
	mux.HandleFunc("GET /auth/me", h.AuthMiddleware(h.HandleGetCurrentUser))
	mux.HandleFunc("PUT /auth/me", h.AuthMiddleware(h.HandleUpdateProfile))
	mux.HandleFunc("PUT /auth/change-password", h.AuthMiddleware(h.HandleChangePassword))
}

// HandleRegister handles user registration
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reg UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract tenant ID from request
	// In a real app, this might come from a subdomain or other mechanism
	// For now, we'll use a query parameter for demonstration
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default" // Use a default tenant for simplicity
	}

	// Register the user
	user, err := h.service.Register(reg, tenantID)
	if err != nil {
		if err.Error() == "email already in use" {
			sendJSONError(w, "Email already in use", http.StatusConflict)
			return
		}
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// HandleLogin handles user login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var login UserLogin
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	response, err := h.service.Login(login)
	if err != nil {
		// Use a generic error message for security
		sendJSONError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Return the login response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleRefreshToken handles token refresh
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Refresh the token
	token, expiresAt, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		sendJSONError(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Return the new access token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// HandleGetCurrentUser handles retrieving the current user's profile
func (h *Handler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user by ID
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the user profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleUpdateProfile handles updating the current user's profile
func (h *Handler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var profile UserProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the profile
	user, err := h.service.UpdateUserProfile(userID, profile)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			sendJSONError(w, "Email already in use", http.StatusConflict)
			return
		}
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleChangePassword handles password change
func (h *Handler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		sendJSONError(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Change the password
	err := h.service.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			sendJSONError(w, "Current password is incorrect", http.StatusUnauthorized)
			return
		}
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

// AuthMiddleware authenticates requests using JWT tokens
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendJSONError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from bearer prefix
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sendJSONError(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}
		tokenString := tokenParts[1]

		// Validate token
		claims, err := h.service.ValidateToken(tokenString)
		if err != nil {
			sendJSONError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = NewContextWithUserID(ctx, claims.UserID)
		ctx = NewContextWithTenantID(ctx, claims.TenantID)
		ctx = NewContextWithUserRole(ctx, string(claims.Role))

		// Call the next handler with the updated context
		next(w, r.WithContext(ctx))
	}
}

// TenantAuthMiddleware ensures the user belongs to the specified tenant
func (h *Handler) TenantAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return h.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant ID from context
		tenantID, ok := GetTenantIDFromContext(r.Context())
		if !ok {
			sendJSONError(w, "Tenant ID not found in token", http.StatusUnauthorized)
			return
		}

		// Get tenant ID from request (e.g., from path parameter or query)
		requestTenantID := r.URL.Query().Get("tenant_id")
		if requestTenantID == "" {
			sendJSONError(w, "Tenant ID required", http.StatusBadRequest)
			return
		}

		// Ensure tenant IDs match
		if tenantID != requestTenantID {
			sendJSONError(w, "Access denied to this tenant", http.StatusForbidden)
			return
		}

		// Call the next handler
		next(w, r)
	})
}

// AdminAuthMiddleware ensures the user has admin role
func (h *Handler) AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return h.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get role from context
		role, ok := GetUserRoleFromContext(r.Context())
		if !ok {
			sendJSONError(w, "Role not found in token", http.StatusUnauthorized)
			return
		}

		// Ensure user is an admin
		if role != string(RoleAdmin) {
			sendJSONError(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Call the next handler
		next(w, r)
	})
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ResponseError{Error: message})
}
