package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/joshbarros/golang-carflow-api/internal/auth"
	"github.com/joshbarros/golang-carflow-api/internal/cache"
	"github.com/joshbarros/golang-carflow-api/internal/car"
	"github.com/joshbarros/golang-carflow-api/internal/database"
	"github.com/joshbarros/golang-carflow-api/internal/metrics"
	"github.com/joshbarros/golang-carflow-api/internal/middleware"
	"github.com/joshbarros/golang-carflow-api/internal/tenant"
)

var (
	// Global cache instance
	globalCache *cache.Cache
)

func main() {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load() // ignore error if file doesn't exist

	// Get environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database configuration
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "carflow"),
		Password: getEnv("DB_PASSWORD", "carflow_secret"),
		DBName:   getEnv("DB_NAME", "carflow"),
	}

	// Initialize database connection
	db, err := database.New(dbConfig)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		os.Exit(1)
	}

	// Run database migrations
	n, err := database.MigrateUp(db)
	if err != nil {
		log.Printf("Failed to run database migrations: %v", err)
		os.Exit(1)
	}
	log.Printf("Applied %d database migrations", n)

	// Initialize repositories
	authRepo, err := auth.NewPostgresRepository(db)
	if err != nil {
		log.Printf("Failed to create auth repository: %v", err)
		os.Exit(1)
	}

	carRepo, err := car.NewPostgresRepository(db.DB)
	if err != nil {
		log.Printf("Failed to create car repository: %v", err)
		os.Exit(1)
	}

	tenantRepo, err := tenant.NewPostgresRepository(db.DB)
	if err != nil {
		log.Printf("Failed to create tenant repository: %v", err)
		os.Exit(1)
	}

	// Create services
	authService := auth.NewService(authRepo)
	carService := car.NewService(carRepo)
	tenantService := tenant.NewService(tenantRepo)

	// Create handlers
	authHandler := auth.NewHandler(authService)
	carHandler := car.NewHandler(carService)

	// Create the metrics tracker
	metricsTracker := metrics.NewMetrics()

	// Create router
	mux := http.NewServeMux()

	// Register auth routes
	authHandler.RegisterRoutes(mux)

	// Register car routes with auth middleware
	mux.HandleFunc("GET /cars", authHandler.AuthMiddleware(carHandler.HandleGetAllCars))
	mux.HandleFunc("GET /cars/{id}", authHandler.AuthMiddleware(carHandler.HandleGetCar))
	mux.HandleFunc("POST /cars", authHandler.AuthMiddleware(carHandler.HandleCreateCar))
	mux.HandleFunc("PUT /cars/{id}", authHandler.AuthMiddleware(carHandler.HandleUpdateCar))
	mux.HandleFunc("DELETE /cars/{id}", authHandler.AuthMiddleware(carHandler.HandleDeleteCar))

	// Add health check endpoint
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add API docs endpoint
	mux.HandleFunc("GET /api-docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.json")
	})

	// Add Swagger UI endpoints
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/swagger-ui/index.html")
	})

	// Serve Swagger UI static files
	mux.HandleFunc("GET /swagger-ui/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/swagger-ui/"):]
		http.ServeFile(w, r, "public/swagger-ui/"+path)
	})

	// Create rate limiter with tenant service
	rateLimiter := middleware.NewRateLimiter(tenantService)

	// Create tenant context middleware
	tenantContextMiddleware := middleware.NewTenantContextMiddleware(db.DB)

	// Create a chain of middlewares
	handler := middleware.LoggingMiddleware(
		middleware.CORSMiddleware(
			middleware.RecoveryMiddleware(
				middleware.RateLimitMiddleware(rateLimiter)(
					tenantContextMiddleware.Middleware(
						middleware.ETagMiddleware(
							metrics.MetricsMiddleware(metricsTracker)(mux),
						),
					),
				),
			),
		),
	)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// seedData adds sample data to the repositories
func seedData(carService *car.Service, authService *auth.Service) {
	// Add sample cars
	sampleCars := []car.Car{
		{ID: "1", Make: "Toyota", Model: "Corolla", Year: 2020, Color: "blue"},
		{ID: "2", Make: "Honda", Model: "Civic", Year: 2019, Color: "red"},
		{ID: "3", Make: "Tesla", Model: "Model 3", Year: 2022, Color: "white"},
	}

	for _, c := range sampleCars {
		_, err := carService.CreateCar(c)
		if err != nil {
			log.Printf("Error seeding car data: %v", err)
		}
	}

	// Add sample users
	adminUser := auth.UserRegistration{
		Email:     "admin@carflow.com",
		Password:  "adminpass123",
		FirstName: "Admin",
		LastName:  "User",
	}

	regularUser := auth.UserRegistration{
		Email:     "user@carflow.com",
		Password:  "userpass123",
		FirstName: "Regular",
		LastName:  "User",
	}

	// Create an admin user
	admin, err := authService.Register(adminUser, "default")
	if err != nil {
		log.Printf("Error seeding admin user: %v", err)
	} else {
		// Update the role to admin
		admin.Role = auth.RoleAdmin
		_, err = authService.UpdateUserProfile(admin.ID, auth.UserProfile{
			Email:     admin.Email,
			FirstName: admin.FirstName,
			LastName:  admin.LastName,
		})
		if err != nil {
			log.Printf("Error updating admin role: %v", err)
		}
	}

	// Create a regular user
	_, err = authService.Register(regularUser, "default")
	if err != nil {
		log.Printf("Error seeding regular user: %v", err)
	}

	log.Println("Sample data loaded")
}

// Test comment
// Trigger build
