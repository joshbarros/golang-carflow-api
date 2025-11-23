package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, databaseURL string
	var direction, steps int

	flag.StringVar(&migrationsPath, "path", "./migrations", "Path to migrations directory")
	flag.StringVar(&databaseURL, "database", getEnv("DATABASE_URL", ""), "Database connection string")
	flag.IntVar(&direction, "direction", 1, "Migration direction: 1 for up, -1 for down")
	flag.IntVar(&steps, "steps", 0, "Number of steps to migrate (0 = all)")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required. Set it via environment variable or -database flag")
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("Failed to get current version: %v", err)
	}

	if dirty {
		log.Printf("WARNING: Database is in dirty state at version %d", version)
	} else if err == migrate.ErrNilVersion {
		log.Println("Database is not initialized (no version)")
	} else {
		log.Printf("Current database version: %d", version)
	}

	// Run migration
	if direction > 0 {
		if steps > 0 {
			log.Printf("Migrating UP by %d steps...", steps)
			err = m.Steps(steps)
		} else {
			log.Println("Migrating UP to latest version...")
			err = m.Up()
		}
	} else {
		if steps > 0 {
			log.Printf("Migrating DOWN by %d steps...", steps)
			err = m.Steps(-steps)
		} else {
			log.Println("Migrating DOWN all the way...")
			err = m.Down()
		}
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to apply")
	} else {
		// Get new version
		version, _, err = m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			log.Fatalf("Failed to get new version: %v", err)
		}
		if err == migrate.ErrNilVersion {
			log.Println("Successfully migrated! Database version: none")
		} else {
			log.Printf("Successfully migrated! Database version: %d", version)
		}
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
