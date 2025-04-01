package tenant

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/joshbarros/golang-carflow-api/internal/auth"
	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

// Handler handles HTTP requests for tenants
type Handler struct {
	service *Service
}

// NewHandler creates a new tenant handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the tenant routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Admin-only routes
	mux.HandleFunc("GET /admin/tenants", auth.AdminOnlyMiddleware(h.HandleListTenants))
	mux.HandleFunc("POST /admin/tenants", auth.AdminOnlyMiddleware(h.HandleCreateTenant))
	mux.HandleFunc("GET /admin/tenants/{id}", auth.AdminOnlyMiddleware(h.HandleGetTenant))
	mux.HandleFunc("PUT /admin/tenants/{id}", auth.AdminOnlyMiddleware(h.HandleUpdateTenant))
	mux.HandleFunc("DELETE /admin/tenants/{id}", auth.AdminOnlyMiddleware(h.HandleDeleteTenant))
}

// HandleListTenants handles GET /admin/tenants
func (h *Handler) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get tenants
	tenants, err := h.service.ListTenants(page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenants)
}

// HandleCreateTenant handles POST /admin/tenants
func (h *Handler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var tenant domain.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create tenant
	if err := h.service.CreateTenant(tenant); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get created tenant
	createdTenant, err := h.service.GetTenant(tenant.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTenant)
}

// HandleGetTenant handles GET /admin/tenants/{id}
func (h *Handler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from URL
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	// Get tenant
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenant)
}

// HandleUpdateTenant handles PUT /admin/tenants/{id}
func (h *Handler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from URL
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var tenant domain.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches body
	if tenant.ID != id {
		http.Error(w, "Tenant ID mismatch", http.StatusBadRequest)
		return
	}

	// Update tenant
	if err := h.service.UpdateTenant(tenant); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get updated tenant
	updatedTenant, err := h.service.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTenant)
}

// HandleDeleteTenant handles DELETE /admin/tenants/{id}
func (h *Handler) HandleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from URL
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	// Delete tenant
	if err := h.service.DeleteTenant(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}
