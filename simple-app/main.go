package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Car represents a car entity
type Car struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Color string `json:"color"`
}

// Response represents the API response format
type Response struct {
	Data       []Car `json:"data"`
	TotalItems int   `json:"total_items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

// Track server start time
var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)

	// Define handlers
	http.HandleFunc("/cars", handleCars)
	http.HandleFunc("/healthz", handleHealth)

	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleCars(w http.ResponseWriter, r *http.Request) {
	// Sample cars data
	cars := []Car{
		{ID: "1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
		{ID: "2", Make: "Honda", Model: "Civic", Year: 2019, Color: "red"},
		{ID: "3", Make: "Tesla", Model: "Model 3", Year: 2022, Color: "white"},
	}

	// Create response
	response := Response{
		Data:       cars,
		TotalItems: len(cars),
		Page:       1,
		PageSize:   10,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Health check response
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    fmt.Sprintf("%s", time.Since(startTime)),
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
