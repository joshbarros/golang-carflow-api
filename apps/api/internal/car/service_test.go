package car

import (
	"strings"
	"testing"
)

func TestValidateCar(t *testing.T) {
	tests := []struct {
		name    string
		car     Car
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid car",
			car:     Car{ID: "test1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
			wantErr: false,
		},
		{
			name:    "Empty ID",
			car:     Car{ID: "", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
			wantErr: true,
			errMsg:  "ID is required",
		},
		{
			name:    "Invalid ID",
			car:     Car{ID: "test@123", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
			wantErr: true,
			errMsg:  "ID must be alphanumeric",
		},
		{
			name:    "Empty Make",
			car:     Car{ID: "test1", Make: "", Model: "Corolla", Year: 2020, Color: "blue"},
			wantErr: true,
			errMsg:  "make is required",
		},
		{
			name:    "Empty Model",
			car:     Car{ID: "test1", Make: "Toyota", Model: "", Year: 2020, Color: "blue"},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name:    "Year too old",
			car:     Car{ID: "test1", Make: "Toyota", Model: "Corolla", Year: 1800, Color: "blue"},
			wantErr: true,
			errMsg:  "year must be between",
		},
		{
			name:    "Year too new",
			car:     Car{ID: "test1", Make: "Toyota", Model: "Corolla", Year: 3001, Color: "blue"},
			wantErr: true,
			errMsg:  "year must be between",
		},
		{
			name:    "Invalid color",
			car:     Car{ID: "test1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue@123"},
			wantErr: true,
			errMsg:  "color must be alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCar(tt.car)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateCar() error = %v, expected to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestService_GetCar(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	// Add a test car
	testCar := Car{ID: "service-test-1", Make: "Tesla", Model: "Model S", Year: 2021, Color: "black"}
	repo.Create(testCar)

	// Test retrieval
	car, err := service.GetCar("service-test-1")
	if err != nil {
		t.Errorf("GetCar() error = %v", err)
	}
	if car.ID != testCar.ID || car.Make != testCar.Make {
		t.Errorf("GetCar() = %v, want %v", car, testCar)
	}

	// Test error case
	_, err = service.GetCar("nonexistent")
	if err != ErrNotFound {
		t.Errorf("GetCar() error = %v, want %v", err, ErrNotFound)
	}
}

func TestService_GetAllCars(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	// Empty repository
	cars := service.GetAllCars()
	if len(cars) != 0 {
		t.Errorf("GetAllCars() = %v, want empty slice", cars)
	}

	// Add some cars
	repo.Create(Car{ID: "all-1", Make: "Honda", Model: "Accord", Year: 2019, Color: "silver"})
	repo.Create(Car{ID: "all-2", Make: "Nissan", Model: "Altima", Year: 2020, Color: "white"})

	// Test retrieval
	cars = service.GetAllCars()
	if len(cars) != 2 {
		t.Errorf("GetAllCars() = %v, want 2 cars", len(cars))
	}
}

func TestService_CreateCar(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	// Valid car
	car := Car{ID: "create-1", Make: "Ford", Model: "F-150", Year: 2022, Color: "red"}
	createdCar, err := service.CreateCar(car)
	if err != nil {
		t.Errorf("CreateCar() error = %v", err)
	}
	if createdCar.ID != car.ID {
		t.Errorf("CreateCar() = %v, want %v", createdCar, car)
	}

	// Invalid car
	_, err = service.CreateCar(Car{ID: "", Make: "Ford", Model: "F-150", Year: 2022, Color: "red"})
	if err == nil {
		t.Errorf("CreateCar() expected error for invalid car")
	}
}

func TestService_UpdateCar(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	// Add a car to update
	repo.Create(Car{ID: "update-service-1", Make: "BMW", Model: "X3", Year: 2020, Color: "blue"})

	// Update valid car
	updatedCar := Car{ID: "update-service-1", Make: "BMW", Model: "X3", Year: 2021, Color: "black"}
	result, err := service.UpdateCar(updatedCar)
	if err != nil {
		t.Errorf("UpdateCar() error = %v", err)
	}
	if result.Year != 2021 || result.Color != "black" {
		t.Errorf("UpdateCar() = %v, want %v", result, updatedCar)
	}

	// Invalid car
	_, err = service.UpdateCar(Car{ID: "update-service-1", Make: "", Model: "X3", Year: 2021, Color: "black"})
	if err == nil {
		t.Errorf("UpdateCar() expected error for invalid car")
	}

	// Non-existent car
	_, err = service.UpdateCar(Car{ID: "nonexistent", Make: "BMW", Model: "X3", Year: 2021, Color: "black"})
	if err != ErrNotFound {
		t.Errorf("UpdateCar() error = %v, want %v", err, ErrNotFound)
	}
}
