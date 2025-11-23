package car

import (
	"testing"
)

func TestInMemoryRepository_GetAll(t *testing.T) {
	repo := NewInMemoryRepository()

	// Initially, repository should be empty
	cars := repo.GetAll()
	if len(cars) != 0 {
		t.Errorf("Expected empty repository, got %d cars", len(cars))
	}

	// Add some cars
	repo.Create(Car{ID: "1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"})
	repo.Create(Car{ID: "2", Make: "Honda", Model: "Civic", Year: 2019, Color: "red"})

	// Now we should have 2 cars
	cars = repo.GetAll()
	if len(cars) != 2 {
		t.Errorf("Expected 2 cars, got %d", len(cars))
	}
}

func TestInMemoryRepository_Get(t *testing.T) {
	repo := NewInMemoryRepository()

	// Add a car
	testCar := Car{ID: "test1", Make: "Tesla", Model: "Model 3", Year: 2022, Color: "white"}
	repo.Create(testCar)

	// Test successful retrieval
	car, err := repo.Get("test1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if car.ID != "test1" || car.Make != "Tesla" || car.Model != "Model 3" {
		t.Errorf("Retrieved car doesn't match the original: %v", car)
	}

	// Test non-existent car
	_, err = repo.Get("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for nonexistent car, got %v", err)
	}

	// Test empty ID
	_, err = repo.Get("")
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID for empty ID, got %v", err)
	}
}

func TestInMemoryRepository_Create(t *testing.T) {
	repo := NewInMemoryRepository()

	// Test successful creation
	car, err := repo.Create(Car{ID: "1", Make: "Ford", Model: "Mustang", Year: 2021, Color: "black"})
	if err != nil {
		t.Errorf("Expected no error on create, got %v", err)
	}
	if car.ID != "1" || car.Make != "Ford" {
		t.Errorf("Returned car doesn't match created car: %v", car)
	}

	// Test duplicate ID
	_, err = repo.Create(Car{ID: "1", Make: "Dodge", Model: "Charger", Year: 2020, Color: "green"})
	if err == nil {
		t.Error("Expected error when creating car with duplicate ID, got nil")
	}

	// Test empty ID
	_, err = repo.Create(Car{ID: "", Make: "BMW", Model: "X5", Year: 2022, Color: "silver"})
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID for empty ID, got %v", err)
	}
}

func TestInMemoryRepository_Update(t *testing.T) {
	repo := NewInMemoryRepository()

	// Add a car to update
	repo.Create(Car{ID: "update1", Make: "Audi", Model: "A4", Year: 2020, Color: "gray"})

	// Test successful update
	updatedCar := Car{ID: "update1", Make: "Audi", Model: "A4", Year: 2021, Color: "silver"}
	car, err := repo.Update(updatedCar)
	if err != nil {
		t.Errorf("Expected no error on update, got %v", err)
	}
	if car.Year != 2021 || car.Color != "silver" {
		t.Errorf("Car not properly updated: %v", car)
	}

	// Verify the update by getting the car
	retrievedCar, _ := repo.Get("update1")
	if retrievedCar.Year != 2021 || retrievedCar.Color != "silver" {
		t.Errorf("Updated car not found in repository: %v", retrievedCar)
	}

	// Test non-existent car
	_, err = repo.Update(Car{ID: "nonexistent", Make: "Jeep", Model: "Wrangler", Year: 2022, Color: "green"})
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for nonexistent car, got %v", err)
	}

	// Test empty ID
	_, err = repo.Update(Car{ID: "", Make: "Ferrari", Model: "F8", Year: 2022, Color: "red"})
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID for empty ID, got %v", err)
	}
}
