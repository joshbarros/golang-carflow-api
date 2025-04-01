package car

import "errors"

var (
	// ErrNotFound is returned when a car with the specified ID doesn't exist
	ErrNotFound = errors.New("car not found")
	// ErrInvalidID is returned when an invalid ID is provided
	ErrInvalidID = errors.New("invalid id")
)

// Repository defines the interface for car data access
type Repository interface {
	// Get retrieves a car by ID and tenant_id
	Get(id string, tenantID string) (Car, error)

	// GetAll retrieves all cars for a tenant
	GetAll(tenantID string) []Car

	// Create creates a new car
	Create(car Car) (Car, error)

	// Update updates an existing car
	Update(car Car) (Car, error)

	// Delete removes a car by ID and tenant_id
	Delete(id string, tenantID string) error
}
