package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joshbarros/golang-carflow-api/internal/auth"
	"github.com/joshbarros/golang-carflow-api/internal/cache"
	"github.com/joshbarros/golang-carflow-api/internal/car"
	"github.com/joshbarros/golang-carflow-api/internal/database"
	"github.com/joshbarros/golang-carflow-api/internal/health"
	"github.com/joshbarros/golang-carflow-api/internal/metrics"
	"github.com/joshbarros/golang-carflow-api/internal/middleware"
	"github.com/joshbarros/golang-carflow-api/internal/tenant"
	"github.com/joshbarros/golang-carflow-api/internal/user"
)

var (
	// Global cache instance
	globalCache *cache.Cache
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", getEnvInt("PORT", 8080), "Port to listen on")
	rateLimit := flag.Int("rate-limit", getEnvInt("RATE_LIMIT_RPS", 100), "Rate limit in requests per second")
	rateBurst := flag.Int("rate-burst", 20, "Maximum burst size for rate limiting")
	flag.Parse()

	// Configure logger
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("🚀 Starting CarFlow SaaS API...")

	// Connect to PostgreSQL database
	log.Println("📊 Connecting to PostgreSQL...")
	db, err := database.ConnectFromEnv()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	// Initialize cache
	globalCache = cache.New(5 * time.Minute)

	// Create metrics tracker
	metricsTracker := metrics.NewMetrics()
	metricsHandler := metrics.NewHandler(metricsTracker)

	// Initialize repositories
	userRepo := user.NewPostgresRepository(db)
	tenantRepo := tenant.NewPostgresRepository(db)

	// For demo purposes, we'll use a default tenant ID for car operations
	// In production, this should come from the JWT token
	defaultTenantID := getEnv("DEFAULT_TENANT_ID", "550e8400-e29b-41d4-a716-446655440000")
	carRepo := car.NewPostgresRepository(db, defaultTenantID)

	// Initialize services
	authService := auth.NewService(userRepo, tenantRepo)
	carService := car.NewService(carRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	carHandler := car.NewHandler(carService)
	healthHandler := health.NewHandler()

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(*rateLimit, *rateBurst, 10*time.Minute)

	// Create HTTP server with routes
	mux := http.NewServeMux()

	// ==========================================
	// Public Routes (No authentication required)
	// ==========================================

	// Health & Metrics
	healthHandler.RegisterRoutes(mux)
	metricsHandler.RegisterRoutes(mux)

	// API Documentation
	mux.HandleFunc("GET /api-docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.json")
	})

	// Authentication endpoints (public)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// ==========================================
	// Protected Routes (JWT authentication required)
	// ==========================================

	// Auth protected endpoints
	mux.Handle("GET /api/v1/auth/me", middleware.JWTMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.Handle("POST /api/v1/auth/refresh", middleware.JWTMiddleware(http.HandlerFunc(authHandler.Refresh)))

	// Car endpoints (protected)
	// Note: In production, these should use tenant-specific repos based on JWT token
	mux.Handle("GET /cars", middleware.JWTMiddleware(http.HandlerFunc(carHandler.ListCars)))
	mux.Handle("GET /cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.GetCar)))
	mux.Handle("POST /cars", middleware.JWTMiddleware(http.HandlerFunc(carHandler.CreateCar)))
	mux.Handle("PUT /cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.UpdateCar)))
	mux.Handle("DELETE /cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.DeleteCar)))

	// Also register car endpoints under /api/v1 for consistency
	mux.Handle("GET /api/v1/cars", middleware.JWTMiddleware(http.HandlerFunc(carHandler.ListCars)))
	mux.Handle("GET /api/v1/cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.GetCar)))
	mux.Handle("POST /api/v1/cars", middleware.JWTMiddleware(http.HandlerFunc(carHandler.CreateCar)))
	mux.Handle("PUT /api/v1/cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.UpdateCar)))
	mux.Handle("DELETE /api/v1/cars/{id}", middleware.JWTMiddleware(http.HandlerFunc(carHandler.DeleteCar)))

	// ==========================================
	// Middleware Chain
	// ==========================================

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

	// ==========================================
	// Start Server
	// ==========================================

	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Println("✅ CarFlow SaaS API is ready!")
	log.Println("")
	log.Println("📍 Endpoints:")
	log.Printf("   API:            http://localhost%s", addr)
	log.Printf("   Health:         http://localhost%s/healthz", addr)
	log.Printf("   Metrics:        http://localhost%s/metrics", addr)
	log.Printf("   API Docs:       http://localhost%s/api-docs", addr)
	log.Println("")
	log.Println("🔐 Authentication:")
	log.Printf("   Register:       POST http://localhost%s/api/v1/auth/register", addr)
	log.Printf("   Login:          POST http://localhost%s/api/v1/auth/login", addr)
	log.Printf("   Get User:       GET  http://localhost%s/api/v1/auth/me", addr)
	log.Printf("   Refresh Token:  POST http://localhost%s/api/v1/auth/refresh", addr)
	log.Println("")
	log.Println("🚗 Vehicles (Protected):")
	log.Printf("   List:           GET    http://localhost%s/api/v1/cars", addr)
	log.Printf("   Get:            GET    http://localhost%s/api/v1/cars/{{id}}", addr)
	log.Printf("   Create:         POST   http://localhost%s/api/v1/cars", addr)
	log.Printf("   Update:         PUT    http://localhost%s/api/v1/cars/{{id}}", addr)
	log.Printf("   Delete:         DELETE http://localhost%s/api/v1/cars/{{id}}", addr)
	log.Println("")
	log.Printf("🚀 Server listening on http://localhost%s", addr)
	log.Println("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnv returns an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt returns an environment variable as int or a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}
