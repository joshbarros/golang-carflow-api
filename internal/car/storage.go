package car

import (
	"errors"
	"sync"
)

var (
	// ErrNotFound is returned when a car with the specified ID doesn't exist
	ErrNotFound = errors.New("car not found")
	// ErrInvalidID is returned when an invalid ID is provided
	ErrInvalidID = errors.New("invalid id")
)

// Repository defines the interface for car data access
type Repository interface {
	Get(id string) (Car, error)
	GetAll() []Car
	Create(car Car) (Car, error)
	Update(car Car) (Car, error)
	Delete(id string) error
}

// InMemoryRepository implements Repository interface with an in-memory data store
type InMemoryRepository struct {
	cars map[string]Car
	mu   sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		cars: make(map[string]Car),
	}
}

// Get retrieves a car by ID
func (r *InMemoryRepository) Get(id string) (Car, error) {
	if id == "" {
		return Car{}, ErrInvalidID
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	car, ok := r.cars[id]
	if !ok {
		return Car{}, ErrNotFound
	}
	return car, nil
}

// GetAll retrieves all cars
func (r *InMemoryRepository) GetAll() []Car {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cars := make([]Car, 0, len(r.cars))
	for _, car := range r.cars {
		cars = append(cars, car)
	}
	return cars
}

// Create adds a new car to the repository
func (r *InMemoryRepository) Create(car Car) (Car, error) {
	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if car already exists
	if _, exists := r.cars[car.ID]; exists {
		return Car{}, errors.New("car with this ID already exists")
	}

	r.cars[car.ID] = car
	return car, nil
}

// Update updates an existing car
func (r *InMemoryRepository) Update(car Car) (Car, error) {
	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if car exists
	if _, exists := r.cars[car.ID]; !exists {
		return Car{}, ErrNotFound
	}

	r.cars[car.ID] = car
	return car, nil
}

// Delete removes a car from the repository
func (r *InMemoryRepository) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if car exists
	if _, exists := r.cars[id]; !exists {
		return ErrNotFound
	}

	delete(r.cars, id)
	return nil
}
