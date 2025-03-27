package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joshbarros/golang-carflow-api/internal/car"
	"github.com/joshbarros/golang-carflow-api/internal/health"
	"github.com/joshbarros/golang-carflow-api/internal/metrics"
	"github.com/joshbarros/golang-carflow-api/internal/middleware"
)

func setupTestServer() *httptest.Server {
	// Create components
	metricsTracker := metrics.NewMetrics()
	metricsHandler := metrics.NewHandler(metricsTracker)

	carRepo := car.NewInMemoryRepository()
	carService := car.NewService(carRepo)
	carHandler := car.NewHandler(carService)

	healthHandler := health.NewHandler()

	// Add some sample data
	carService.CreateCar(car.Car{ID: "test1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"})

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

func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	resp, err := http.Get(fmt.Sprintf("%s/healthz", server.URL))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status, ok := result["status"].(string); !ok || status != "ok" {
		t.Errorf("Expected status to be 'ok', got %v", result["status"])
	}
}

// PagedResponse represents the paginated response structure
type PagedResponse struct {
	Data     []car.Car `json:"data"`
	Metadata struct {
		Total    int `json:"total"`
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Pages    int `json:"pages"`
	} `json:"metadata"`
}

func TestCarsEndpoints(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Get All Cars", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/cars", server.URL))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		// Print response for debugging
		t.Logf("Response body: %s", string(bodyBytes))

		// Try to parse as a generic map first to inspect structure
		var responseMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
			t.Fatalf("Failed to parse response as JSON: %v", err)
		}

		// Now properly validate the response format
		data, hasData := responseMap["data"]
		if !hasData {
			t.Fatalf("Response doesn't contain 'data' field")
		}

		cars, ok := data.([]interface{})
		if !ok {
			t.Fatalf("Data is not an array")
		}

		if len(cars) == 0 {
			t.Error("Expected at least one car, got none")
		}

		// Check total_items field
		totalItems, hasTotalItems := responseMap["total_items"]
		if !hasTotalItems {
			t.Fatalf("Response doesn't contain 'total_items' field")
		}

		totalItemsValue, ok := totalItems.(float64)
		if !ok {
			t.Fatalf("total_items is not a number, got %T", totalItems)
		}

		if totalItemsValue < 1 {
			t.Errorf("Expected total_items to be at least 1, got %v", totalItemsValue)
		}

		// Check page field
		page, hasPage := responseMap["page"]
		if !hasPage {
			t.Fatalf("Response doesn't contain 'page' field")
		}

		pageValue, ok := page.(float64)
		if !ok {
			t.Fatalf("page is not a number, got %T", page)
		}

		if pageValue < 1 {
			t.Errorf("Expected page to be at least 1, got %v", pageValue)
		}

		// Check page_size field
		pageSize, hasPageSize := responseMap["page_size"]
		if !hasPageSize {
			t.Fatalf("Response doesn't contain 'page_size' field")
		}

		pageSizeValue, ok := pageSize.(float64)
		if !ok {
			t.Fatalf("page_size is not a number, got %T", pageSize)
		}

		if pageSizeValue < 1 {
			t.Errorf("Expected page_size to be at least 1, got %v", pageSizeValue)
		}
	})

	t.Run("Get Single Car", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/cars/test1", server.URL))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var car car.Car
		if err := json.NewDecoder(resp.Body).Decode(&car); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if car.ID != "test1" || car.Make != "Toyota" || car.Model != "Corolla" {
			t.Errorf("Car data does not match expected: %+v", car)
		}
	})

	t.Run("Create Car", func(t *testing.T) {
		newCar := car.Car{
			ID:    "test-create",
			Make:  "Honda",
			Model: "Civic",
			Year:  2021,
			Color: "red",
		}

		payload, err := json.Marshal(newCar)
		if err != nil {
			t.Fatalf("Failed to marshal car: %v", err)
		}

		resp, err := http.Post(
			fmt.Sprintf("%s/cars", server.URL),
			"application/json",
			bytes.NewBuffer(payload),
		)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var createdCar car.Car
		if err := json.NewDecoder(resp.Body).Decode(&createdCar); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if createdCar.ID != newCar.ID || createdCar.Make != newCar.Make || createdCar.Model != newCar.Model {
			t.Errorf("Created car does not match expected: %+v", createdCar)
		}
	})

	t.Run("Update Car", func(t *testing.T) {
		updatedCar := car.Car{
			ID:    "test1",
			Make:  "Toyota",
			Model: "Corolla",
			Year:  2022,
			Color: "green",
		}

		payload, err := json.Marshal(updatedCar)
		if err != nil {
			t.Fatalf("Failed to marshal car: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("%s/cars/test1", server.URL),
			bytes.NewBuffer(payload),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var resultCar car.Car
		if err := json.NewDecoder(resp.Body).Decode(&resultCar); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resultCar.Year != 2022 || resultCar.Color != "green" {
			t.Errorf("Updated car does not match expected: %+v", resultCar)
		}
	})
}

func TestMain(m *testing.M) {
	// Setup
	os.Exit(m.Run())
}
