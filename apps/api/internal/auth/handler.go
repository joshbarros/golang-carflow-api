package auth

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Register(&req)
	if err != nil {
		if err == ErrEmailAlreadyExists {
			respondError(w, err.Error(), http.StatusConflict)
			return
		}
		if err == ErrInvalidEmail || err == ErrPasswordTooShort {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Registration error: %v", err)
		respondError(w, "Failed to register", http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp, http.StatusCreated)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		if err == ErrInvalidCredentials {
			respondError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		if err == ErrUserNotActive {
			respondError(w, "Account is not active", http.StatusForbidden)
			return
		}
		log.Printf("Login error: %v", err)
		respondError(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp, http.StatusOK)
}

// Me handles GET /api/v1/auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		respondError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	respondJSON(w, user, http.StatusOK)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user data from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email, ok := r.Context().Value("email").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, ok := r.Context().Value("role").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate new token
	token, err := GenerateToken(userID, tenantID, email, role)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		respondError(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": token,
	}

	respondJSON(w, response, http.StatusOK)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	respondJSON(w, response, statusCode)
}
