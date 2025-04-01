package car

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRepository implements Repository interface with PostgreSQL storage
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Get retrieves a car by ID and tenant_id for proper isolation
func (r *PostgresRepository) Get(id string, tenantID string) (Car, error) {
	var car Car
	query := `
		SELECT id, make, model, year, color, tenant_id, created_at, updated_at
		FROM cars
		WHERE id = $1 AND tenant_id = $2`

	err := r.db.QueryRow(query, id, tenantID).Scan(
		&car.ID,
		&car.Make,
		&car.Model,
		&car.Year,
		&car.Color,
		&car.TenantID,
		&car.CreatedAt,
		&car.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return Car{}, ErrNotFound
	}
	if err != nil {
		return Car{}, fmt.Errorf("error getting car: %v", err)
	}

	return car, nil
}

// GetAll retrieves all cars for a specific tenant
func (r *PostgresRepository) GetAll(tenantID string) []Car {
	query := `
		SELECT id, make, model, year, color, tenant_id, created_at, updated_at
		FROM cars
		WHERE tenant_id = $1`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var cars []Car
	for rows.Next() {
		var car Car
		err := rows.Scan(
			&car.ID,
			&car.Make,
			&car.Model,
			&car.Year,
			&car.Color,
			&car.TenantID,
			&car.CreatedAt,
			&car.UpdatedAt,
		)
		if err != nil {
			continue
		}
		cars = append(cars, car)
	}

	return cars
}

// Create creates a new car with tenant isolation
func (r *PostgresRepository) Create(car Car) (Car, error) {
	// Validate tenant_id is not empty
	if car.TenantID == "" {
		return Car{}, fmt.Errorf("tenant_id is required")
	}

	query := `
		INSERT INTO cars (id, make, model, year, color, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, make, model, year, color, tenant_id, created_at, updated_at`

	now := time.Now()
	car.CreatedAt = now
	car.UpdatedAt = now

	err := r.db.QueryRow(
		query,
		car.ID,
		car.Make,
		car.Model,
		car.Year,
		car.Color,
		car.TenantID,
		car.CreatedAt,
		car.UpdatedAt,
	).Scan(
		&car.ID,
		&car.Make,
		&car.Model,
		&car.Year,
		&car.Color,
		&car.TenantID,
		&car.CreatedAt,
		&car.UpdatedAt,
	)

	if err != nil {
		return Car{}, fmt.Errorf("error creating car: %v", err)
	}

	return car, nil
}

// Update updates an existing car with tenant isolation
func (r *PostgresRepository) Update(car Car) (Car, error) {
	query := `
		UPDATE cars
		SET make = $1, model = $2, year = $3, color = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7
		RETURNING id, make, model, year, color, tenant_id, created_at, updated_at`

	car.UpdatedAt = time.Now()

	err := r.db.QueryRow(
		query,
		car.Make,
		car.Model,
		car.Year,
		car.Color,
		car.UpdatedAt,
		car.ID,
		car.TenantID,
	).Scan(
		&car.ID,
		&car.Make,
		&car.Model,
		&car.Year,
		&car.Color,
		&car.TenantID,
		&car.CreatedAt,
		&car.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return Car{}, ErrNotFound
	}
	if err != nil {
		return Car{}, fmt.Errorf("error updating car: %v", err)
	}

	return car, nil
}

// Delete removes a car by ID with tenant isolation
func (r *PostgresRepository) Delete(id string, tenantID string) error {
	query := `DELETE FROM cars WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.Exec(query, id, tenantID)
	if err != nil {
		return fmt.Errorf("error deleting car: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
