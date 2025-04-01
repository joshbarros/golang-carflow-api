package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joshbarros/golang-carflow-api/internal/billing"
	"github.com/joshbarros/golang-carflow-api/internal/tenant"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Stripe service
	stripeService, err := billing.NewStripeService()
	if err != nil {
		log.Fatalf("Failed to initialize Stripe service: %v", err)
	}

	// Initialize tenant repository
	tenantRepo, err := tenant.NewPostgresRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize tenant repository: %v", err)
	}

	// Initialize tenant service
	tenantService := tenant.NewService(tenantRepo)

	// Initialize billing handler
	billingHandler := billing.NewHandler(stripeService, tenantService)

	// Create a new mux router
	mux := http.NewServeMux()

	// Register routes
	billingHandler.RegisterRoutes(mux)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
