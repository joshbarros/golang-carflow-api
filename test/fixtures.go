package test

import (
	"github.com/joshbarros/golang-carflow-api/internal/car"
)

// TestCars contains sample cars for testing
var TestCars = []car.Car{
	{
		ID:    "test-car-1",
		Make:  "Toyota",
		Model: "Corolla",
		Year:  2020,
		Color: "blue",
	},
	{
		ID:    "test-car-2",
		Make:  "Honda",
		Model: "Civic",
		Year:  2019,
		Color: "red",
	},
	{
		ID:    "test-car-3",
		Make:  "Tesla",
		Model: "Model 3",
		Year:  2022,
		Color: "white",
	},
	{
		ID:    "test-car-4",
		Make:  "Ford",
		Model: "Mustang",
		Year:  2021,
		Color: "black",
	},
	{
		ID:    "test-car-5",
		Make:  "Chevrolet",
		Model: "Camaro",
		Year:  2023,
		Color: "yellow",
	},
}

// LoadFixtures populates a repository with test cars
func LoadFixtures(repo *car.InMemoryRepository) {
	for _, c := range TestCars {
		repo.Create(c)
	}
}

// GetTestCar returns a specific test car by index
func GetTestCar(index int) car.Car {
	if index < 0 || index >= len(TestCars) {
		return car.Car{}
	}
	return TestCars[index]
}

// CreateTestCar creates a new car for testing
func CreateTestCar(id, make, model string, year int, color string) car.Car {
	return car.Car{
		ID:    id,
		Make:  make,
		Model: model,
		Year:  year,
		Color: color,
	}
}
