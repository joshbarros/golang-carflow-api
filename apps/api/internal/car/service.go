package car

import (
	"errors"
	"regexp"
	"sort"
	"strings"
)

var (
	// ErrInvalidCar is returned when car data is invalid
	ErrInvalidCar = errors.New("invalid car data")
	// ErrIDGeneration is returned when an ID couldn't be generated
	ErrIDGeneration = errors.New("failed to generate ID")
)

// FilterOptions contains options for filtering cars
type FilterOptions struct {
	Make  string
	Model string
	Year  int
	Color string
}

// SortOptions contains options for sorting cars
type SortOptions struct {
	Field string
	Order string // "asc" or "desc"
}

// PaginationOptions contains options for paginating results
type PaginationOptions struct {
	Page     int
	PageSize int
}

// PagedResult represents a paginated result set
type PagedResult struct {
	Data       []Car `json:"data"`
	TotalItems int   `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

// Service handles car business logic
type Service struct {
	repo Repository
}

// NewService creates a new car service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetCar retrieves a car by ID
func (s *Service) GetCar(id string) (Car, error) {
	return s.repo.Get(id)
}

// GetAllCars retrieves all cars
func (s *Service) GetAllCars() []Car {
	return s.repo.GetAll()
}

// GetFilteredCars retrieves cars with filtering and sorting
func (s *Service) GetFilteredCars(filter FilterOptions, sort *SortOptions) []Car {
	// Get all cars
	cars := s.repo.GetAll()

	// Apply filters
	cars = applyFilters(cars, filter)

	// Apply sorting if requested
	if sort != nil && sort.Field != "" {
		cars = applySorting(cars, *sort)
	}

	return cars
}

// GetPagedCars retrieves cars with filtering, sorting, and pagination
func (s *Service) GetPagedCars(filter FilterOptions, sort *SortOptions, pagination PaginationOptions) PagedResult {
	// Get filtered and sorted cars
	filteredCars := s.GetFilteredCars(filter, sort)

	// Total items and pages
	totalItems := len(filteredCars)

	// Default pagination values if not set
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	// Calculate total pages
	totalPages := (totalItems + pagination.PageSize - 1) / pagination.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Ensure page is within bounds
	if pagination.Page > totalPages {
		pagination.Page = totalPages
	}

	// Calculate start and end indices
	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize

	// Ensure end index doesn't exceed array bounds
	if endIndex > totalItems {
		endIndex = totalItems
	}

	// Get the slice of cars for the current page
	var pagedCars []Car
	if startIndex < totalItems {
		pagedCars = filteredCars[startIndex:endIndex]
	} else {
		pagedCars = []Car{}
	}

	return PagedResult{
		Data:       pagedCars,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
	}
}

// CreateCar creates a new car, validating the data
func (s *Service) CreateCar(car Car) (Car, error) {
	if err := validateCar(car); err != nil {
		return Car{}, err
	}

	return s.repo.Create(car)
}

// UpdateCar updates an existing car, validating the data
func (s *Service) UpdateCar(car Car) (Car, error) {
	if err := validateCar(car); err != nil {
		return Car{}, err
	}

	return s.repo.Update(car)
}

// DeleteCar deletes a car by ID
func (s *Service) DeleteCar(id string) error {
	return s.repo.Delete(id)
}

// validateCar checks if car data is valid
func validateCar(car Car) error {
	// ID must be present and in a valid format
	if car.ID == "" {
		return errors.New("ID is required")
	}

	// ID should be alphanumeric, allow dashes and underscores
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, car.ID)
	if !match {
		return errors.New("ID must be alphanumeric, dashes and underscores allowed")
	}

	// Make must be present
	if car.Make == "" {
		return errors.New("make is required")
	}

	// Model must be present
	if car.Model == "" {
		return errors.New("model is required")
	}

	// Year validation
	if car.Year < 1886 || car.Year > 3000 {
		return errors.New("year must be between 1886 and 3000")
	}

	// Color is optional, but should be valid if provided
	if car.Color != "" {
		match, _ = regexp.MatchString(`^[a-zA-Z0-9 ]+$`, car.Color)
		if !match {
			return errors.New("color must be alphanumeric")
		}
	}

	return nil
}

// applyFilters filters the cars based on filter options
func applyFilters(cars []Car, filter FilterOptions) []Car {
	var result []Car

	for _, car := range cars {
		// Check all filters
		if (filter.Make == "" || strings.EqualFold(car.Make, filter.Make)) &&
			(filter.Model == "" || strings.EqualFold(car.Model, filter.Model)) &&
			(filter.Year == 0 || car.Year == filter.Year) &&
			(filter.Color == "" || strings.EqualFold(car.Color, filter.Color)) {
			result = append(result, car)
		}
	}

	return result
}

// applySorting sorts the cars based on sort options
func applySorting(cars []Car, sortOpt SortOptions) []Car {
	result := make([]Car, len(cars))
	copy(result, cars)

	// Determine sort order
	isAscending := sortOpt.Order == "" || strings.ToLower(sortOpt.Order) == "asc"

	// Sort based on field
	switch strings.ToLower(sortOpt.Field) {
	case "make":
		sort.Slice(result, func(i, j int) bool {
			if isAscending {
				return strings.ToLower(result[i].Make) < strings.ToLower(result[j].Make)
			}
			return strings.ToLower(result[i].Make) > strings.ToLower(result[j].Make)
		})
	case "model":
		sort.Slice(result, func(i, j int) bool {
			if isAscending {
				return strings.ToLower(result[i].Model) < strings.ToLower(result[j].Model)
			}
			return strings.ToLower(result[i].Model) > strings.ToLower(result[j].Model)
		})
	case "year":
		sort.Slice(result, func(i, j int) bool {
			if isAscending {
				return result[i].Year < result[j].Year
			}
			return result[i].Year > result[j].Year
		})
	case "color":
		sort.Slice(result, func(i, j int) bool {
			if isAscending {
				return strings.ToLower(result[i].Color) < strings.ToLower(result[j].Color)
			}
			return strings.ToLower(result[i].Color) > strings.ToLower(result[j].Color)
		})
	case "id":
		sort.Slice(result, func(i, j int) bool {
			if isAscending {
				return strings.ToLower(result[i].ID) < strings.ToLower(result[j].ID)
			}
			return strings.ToLower(result[i].ID) > strings.ToLower(result[j].ID)
		})
	}

	return result
}
