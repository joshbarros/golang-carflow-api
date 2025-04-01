package car

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/joshbarros/golang-carflow-api/internal/middleware"
)

// Handler handles HTTP requests for car endpoints
type Handler struct {
	service *Service
}

// NewHandler creates a new car handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the car endpoints to the given ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /cars", h.HandleGetAllCars)
	mux.HandleFunc("GET /cars/{id}", h.HandleGetCar)
	mux.HandleFunc("POST /cars", h.HandleCreateCar)
	mux.HandleFunc("PUT /cars/{id}", h.HandleUpdateCar)
	mux.HandleFunc("DELETE /cars/{id}", h.HandleDeleteCar)
}

// HandleGetAllCars handles GET /cars requests
func (h *Handler) HandleGetAllCars(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(middleware.TenantIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract query parameters for filtering
	query := r.URL.Query()

	// Build filter options
	filter := FilterOptions{
		Make:  query.Get("make"),
		Model: query.Get("model"),
		Color: query.Get("color"),
	}

	// Parse year if provided
	if yearStr := query.Get("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year parameter", http.StatusBadRequest)
			return
		}
		filter.Year = year
	}

	// Extract sorting parameters
	var sortOptions *SortOptions
	if sortField := query.Get("sort"); sortField != "" {
		// Check if sort order is specified
		order := "asc"
		if sortField[0] == '-' {
			order = "desc"
			sortField = sortField[1:]
		}

		// Validate sort field
		validFields := map[string]bool{
			"id":    true,
			"make":  true,
			"model": true,
			"year":  true,
			"color": true,
		}

		if !validFields[sortField] {
			http.Error(w, "Invalid sort field", http.StatusBadRequest)
			return
		}

		sortOptions = &SortOptions{
			Field: sortField,
			Order: order,
		}
	}

	// Extract pagination parameters
	pagination := PaginationOptions{
		Page:     1,
		PageSize: 10,
	}

	// Parse page parameter
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, "Invalid page parameter", http.StatusBadRequest)
			return
		}
		pagination.Page = page
	}

	// Parse page_size parameter
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			http.Error(w, "Invalid page_size parameter (must be between 1 and 100)", http.StatusBadRequest)
			return
		}
		pagination.PageSize = pageSize
	}

	// Check if pagination is requested
	if query.Get("pagination") == "false" {
		// Get cars with filtering and sorting only (no pagination)
		cars := h.service.GetFilteredCars(tenantID, filter, sortOptions)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cars)
	} else {
		// Get cars with filtering, sorting, and pagination
		result := h.service.GetPagedCars(tenantID, filter, sortOptions, pagination)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// HandleGetCar handles GET /cars/{id} requests
func (h *Handler) HandleGetCar(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(middleware.TenantIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/cars/")
	car, err := h.service.GetCar(id, tenantID)

	if err != nil {
		switch err {
		case ErrNotFound:
			http.Error(w, "Car not found", http.StatusNotFound)
		case ErrInvalidID:
			http.Error(w, "Invalid car ID", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(car)
}

// HandleCreateCar handles POST /cars requests
func (h *Handler) HandleCreateCar(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant ID from context
	tenantID, ok := r.Context().Value(middleware.TenantIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set tenant ID
	car.TenantID = tenantID

	// Create car
	createdCar, err := h.service.CreateCar(car)
	if err != nil {
		switch {
		case err == ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		case strings.Contains(err.Error(), "validation"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create car", http.StatusInternalServerError)
		}
		return
	}

	// Return created car
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCar)
}

// HandleUpdateCar handles PUT /cars/{id} requests
func (h *Handler) HandleUpdateCar(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(middleware.TenantIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idPattern := regexp.MustCompile(`/cars/([^/]+)$`)
	matches := idPattern.FindStringSubmatch(r.URL.Path)

	if len(matches) < 2 {
		http.Error(w, "Invalid car ID", http.StatusBadRequest)
		return
	}

	id := matches[1]

	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and tenant ID
	car.ID = id
	car.TenantID = tenantID

	updatedCar, err := h.service.UpdateCar(car)
	if err != nil {
		switch {
		case err == ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		case err == ErrNotFound:
			http.Error(w, "Car not found", http.StatusNotFound)
		case strings.Contains(err.Error(), "validation"):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCar)
}

// HandleDeleteCar handles DELETE /cars/{id} requests
func (h *Handler) HandleDeleteCar(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(middleware.TenantIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idPattern := regexp.MustCompile(`/cars/([^/]+)$`)
	matches := idPattern.FindStringSubmatch(r.URL.Path)

	if len(matches) < 2 {
		http.Error(w, "Invalid car ID", http.StatusBadRequest)
		return
	}

	id := matches[1]

	err := h.service.DeleteCar(id, tenantID)
	if err != nil {
		switch {
		case err == ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		case err == ErrNotFound:
			http.Error(w, "Car not found", http.StatusNotFound)
		case err == ErrInvalidID:
			http.Error(w, "Invalid car ID", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondWithError sends an error response to the client
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response to the client
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
