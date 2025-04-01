package car

import "sync"

// InMemoryRepository implements the Repository interface using an in-memory map
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

// Get retrieves a car by ID and tenant_id
func (r *InMemoryRepository) Get(id string, tenantID string) (Car, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	car, exists := r.cars[id]
	if !exists {
		return Car{}, ErrNotFound
	}

	return car, nil
}

// GetAll retrieves all cars for a tenant
func (r *InMemoryRepository) GetAll(tenantID string) []Car {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var cars []Car
	for _, car := range r.cars {
		cars = append(cars, car)
	}

	return cars
}

// Create creates a new car
func (r *InMemoryRepository) Create(car Car) (Car, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	r.cars[car.ID] = car
	return car, nil
}

// Update updates an existing car
func (r *InMemoryRepository) Update(car Car) (Car, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	if _, exists := r.cars[car.ID]; !exists {
		return Car{}, ErrNotFound
	}

	r.cars[car.ID] = car
	return car, nil
}

// Delete removes a car by ID and tenant_id
func (r *InMemoryRepository) Delete(id string, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.cars[id]; !exists {
		return ErrNotFound
	}

	delete(r.cars, id)
	return nil
}
