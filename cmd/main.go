package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joshbarros/golang-carflow-api/internal/cache"
	"github.com/joshbarros/golang-carflow-api/internal/car"
	"github.com/joshbarros/golang-carflow-api/internal/health"
	"github.com/joshbarros/golang-carflow-api/internal/metrics"
	"github.com/joshbarros/golang-carflow-api/internal/middleware"
)

var (
	// Global cache instance
	globalCache *cache.Cache
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", 8080, "Port to listen on")
	rateLimit := flag.Int("rate-limit", 100, "Rate limit in requests per second")
	rateBurst := flag.Int("rate-burst", 20, "Maximum burst size for rate limiting")
	flag.Parse()

	// Configure logger
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("Starting CarFlow API...")

	// Initialize cache
	globalCache = cache.New(5 * time.Minute) // Cleanup every 5 minutes

	// Create the metrics tracker
	metricsTracker := metrics.NewMetrics()
	metricsHandler := metrics.NewHandler(metricsTracker)

	// Create the car repository and service
	carRepo := car.NewInMemoryRepository()
	carService := car.NewService(carRepo)
	carHandler := car.NewHandler(carService)

	// Create the health check handler
	healthHandler := health.NewHandler()

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(*rateLimit, *rateBurst, 10*time.Minute)

	// Add some sample cars for testing
	seedData(carService)

	// Create the HTTP server
	mux := http.NewServeMux()

	// Register routes
	carHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	metricsHandler.RegisterRoutes(mux)

	// Add API docs endpoint
	mux.HandleFunc("GET /api-docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.json")
	})

	// Create a chain of middlewares
	handler := middleware.CORSMiddleware(
		middleware.RateLimitMiddleware(rateLimiter)(
			middleware.ETagMiddleware(
				metrics.Middleware(metricsTracker)(
					middleware.LoggingMiddleware(
						middleware.RecoveryMiddleware(
							mux,
						),
					),
				),
			),
		),
	)

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("Server listening on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// seedData adds sample cars to the repository
func seedData(service *car.Service) {
	sampleCars := []car.Car{
		{ID: "1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
		{ID: "2", Make: "Honda", Model: "Civic", Year: 2019, Color: "red"},
		{ID: "3", Make: "Tesla", Model: "Model 3", Year: 2022, Color: "white"},
	}

	for _, c := range sampleCars {
		_, err := service.CreateCar(c)
		if err != nil {
			log.Printf("Error seeding car data: %v", err)
		}
	}

	log.Println("Sample car data loaded")
}
