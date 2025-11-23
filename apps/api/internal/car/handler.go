package car

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	mux.HandleFunc("GET /cars", h.handleGetAllCars)
	mux.HandleFunc("GET /cars/{id}", h.handleGetCar)
	mux.HandleFunc("POST /cars", h.handleCreateCar)
	mux.HandleFunc("PUT /cars/{id}", h.handleUpdateCar)
	mux.HandleFunc("DELETE /cars/{id}", h.handleDeleteCar)
}

// handleGetAllCars handles GET /cars requests
func (h *Handler) handleGetAllCars(w http.ResponseWriter, r *http.Request) {
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
			respondWithError(w, http.StatusBadRequest, "Invalid year parameter")
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
			respondWithError(w, http.StatusBadRequest, "Invalid sort field")
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
			respondWithError(w, http.StatusBadRequest, "Invalid page parameter")
			return
		}
		pagination.Page = page
	}

	// Parse page_size parameter
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			respondWithError(w, http.StatusBadRequest, "Invalid page_size parameter (must be between 1 and 100)")
			return
		}
		pagination.PageSize = pageSize
	}

	// Check if pagination is requested
	if query.Get("pagination") == "false" {
		// Get cars with filtering and sorting only (no pagination)
		cars := h.service.GetFilteredCars(filter, sortOptions)
		respondWithJSON(w, http.StatusOK, cars)
	} else {
		// Get cars with filtering, sorting, and pagination
		result := h.service.GetPagedCars(filter, sortOptions, pagination)
		respondWithJSON(w, http.StatusOK, result)
	}
}

// handleGetCar handles GET /cars/{id} requests
func (h *Handler) handleGetCar(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/cars/")
	car, err := h.service.GetCar(id)

	if err != nil {
		switch err {
		case ErrNotFound:
			respondWithError(w, http.StatusNotFound, "Car not found")
		case ErrInvalidID:
			respondWithError(w, http.StatusBadRequest, "Invalid car ID")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, car)
}

// handleCreateCar handles POST /cars requests
func (h *Handler) handleCreateCar(w http.ResponseWriter, r *http.Request) {
	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	createdCar, err := h.service.CreateCar(car)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ID is required") ||
			strings.Contains(err.Error(), "make is required") ||
			strings.Contains(err.Error(), "model is required") ||
			strings.Contains(err.Error(), "year must be between") ||
			strings.Contains(err.Error(), "color must be"):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case strings.Contains(err.Error(), "already exists"):
			respondWithError(w, http.StatusConflict, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, createdCar)
}

// handleUpdateCar handles PUT /cars/{id} requests
func (h *Handler) handleUpdateCar(w http.ResponseWriter, r *http.Request) {
	idPattern := regexp.MustCompile(`/cars/([^/]+)$`)
	matches := idPattern.FindStringSubmatch(r.URL.Path)

	if len(matches) < 2 {
		respondWithError(w, http.StatusBadRequest, "Invalid car ID")
		return
	}

	id := matches[1]

	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Ensure the ID in the URL matches the ID in the body
	car.ID = id

	updatedCar, err := h.service.UpdateCar(car)
	if err != nil {
		switch {
		case err == ErrNotFound:
			respondWithError(w, http.StatusNotFound, "Car not found")
		case strings.Contains(err.Error(), "ID is required") ||
			strings.Contains(err.Error(), "make is required") ||
			strings.Contains(err.Error(), "model is required") ||
			strings.Contains(err.Error(), "year must be between") ||
			strings.Contains(err.Error(), "color must be"):
			respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, updatedCar)
}

// handleDeleteCar handles DELETE /cars/{id} requests
func (h *Handler) handleDeleteCar(w http.ResponseWriter, r *http.Request) {
	idPattern := regexp.MustCompile(`/cars/([^/]+)$`)
	matches := idPattern.FindStringSubmatch(r.URL.Path)

	if len(matches) < 2 {
		respondWithError(w, http.StatusBadRequest, "Invalid car ID")
		return
	}

	id := matches[1]

	err := h.service.DeleteCar(id)
	if err != nil {
		switch err {
		case ErrNotFound:
			respondWithError(w, http.StatusNotFound, "Car not found")
		case ErrInvalidID:
			respondWithError(w, http.StatusBadRequest, "Invalid car ID")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Return 204 No Content on successful deletion
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
