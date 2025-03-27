package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshbarros/golang-carflow-api/internal/car"
	"github.com/joshbarros/golang-carflow-api/internal/health"
	"github.com/joshbarros/golang-carflow-api/internal/metrics"
	"github.com/joshbarros/golang-carflow-api/internal/middleware"
)

// setupBenchmarkServer creates a server for benchmarking
func setupBenchmarkServer() *httptest.Server {
	// Create components
	metricsTracker := metrics.NewMetrics()
	metricsHandler := metrics.NewHandler(metricsTracker)

	carRepo := car.NewInMemoryRepository()
	LoadFixtures(carRepo) // Use our fixtures

	carService := car.NewService(carRepo)
	carHandler := car.NewHandler(carService)

	healthHandler := health.NewHandler()

	// Create server
	mux := http.NewServeMux()

	carHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	metricsHandler.RegisterRoutes(mux)

	// Add middlewares
	handler := metrics.Middleware(metricsTracker)(
		middleware.LoggingMiddleware(
			middleware.RecoveryMiddleware(
				mux,
			),
		),
	)

	return httptest.NewServer(handler)
}

// consumeAndCloseBody reads the response body to completion and closes it
func consumeAndCloseBody(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// BenchmarkGetAllCars benchmarks the GetAll endpoint
func BenchmarkGetAllCars(b *testing.B) {
	server := setupBenchmarkServer()
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/cars", server.URL))
		if err != nil {
			b.Fatalf("Failed to GET /cars: %v", err)
		}
		consumeAndCloseBody(resp)
	}
}

// BenchmarkGetSingleCar benchmarks getting a single car
func BenchmarkGetSingleCar(b *testing.B) {
	server := setupBenchmarkServer()
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/cars/test-car-1", server.URL))
		if err != nil {
			b.Fatalf("Failed to GET /cars/test-car-1: %v", err)
		}
		consumeAndCloseBody(resp)
	}
}

// BenchmarkCreateCar benchmarks creating a new car
func BenchmarkCreateCar(b *testing.B) {
	server := setupBenchmarkServer()
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newCar := CreateTestCar(
			fmt.Sprintf("bench-car-%d", i),
			"Benchmark",
			"TestModel",
			2023,
			"silver",
		)

		payload, err := json.Marshal(newCar)
		if err != nil {
			b.Fatalf("Failed to marshal car: %v", err)
		}

		resp, err := http.Post(
			fmt.Sprintf("%s/cars", server.URL),
			"application/json",
			bytes.NewBuffer(payload),
		)
		if err != nil {
			b.Fatalf("Failed to POST /cars: %v", err)
		}
		consumeAndCloseBody(resp)
	}
}

// BenchmarkUpdateCar benchmarks updating a car
func BenchmarkUpdateCar(b *testing.B) {
	server := setupBenchmarkServer()
	defer server.Close()

	// Create a car to update
	setupCar := CreateTestCar("bench-update-car", "Initial", "Model", 2020, "blue")
	payload, _ := json.Marshal(setupCar)
	resp, err := http.Post(
		fmt.Sprintf("%s/cars", server.URL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		b.Fatalf("Failed to setup test: %v", err)
	}
	consumeAndCloseBody(resp)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		updatedCar := CreateTestCar(
			"bench-update-car",
			"Updated",
			fmt.Sprintf("Model-%d", i),
			2023,
			"red",
		)

		payload, err := json.Marshal(updatedCar)
		if err != nil {
			b.Fatalf("Failed to marshal car: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("%s/cars/bench-update-car", server.URL),
			bytes.NewBuffer(payload),
		)
		if err != nil {
			b.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("Failed to PUT /cars/bench-update-car: %v", err)
		}
		consumeAndCloseBody(resp)
	}
}
