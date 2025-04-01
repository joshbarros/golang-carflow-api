package database

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Config holds the database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// New creates a new database connection
func New(config Config) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// MigrateUp runs all pending migrations
func MigrateUp(db *sqlx.DB) (int, error) {
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrationSource, migrate.Up)
	if err != nil {
		return 0, fmt.Errorf("failed to run migrations: %w", err)
	}

	return n, nil
}

// MigrateDown rolls back all migrations
func MigrateDown(db *sqlx.DB) (int, error) {
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrationSource, migrate.Down)
	if err != nil {
		return 0, fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return n, nil
}
