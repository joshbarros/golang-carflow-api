package car

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PostgresRepository implements Repository interface with PostgreSQL
type PostgresRepository struct {
	db       *sql.DB
	tenantID string // Current tenant context
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB, tenantID string) *PostgresRepository {
	return &PostgresRepository{
		db:       db,
		tenantID: tenantID,
	}
}

// Get retrieves a car by ID (within tenant scope)
func (r *PostgresRepository) Get(id string) (Car, error) {
	if id == "" {
		return Car{}, ErrInvalidID
	}

	query := `
		SELECT id, tenant_id, make, model, year, color, vin, stock_number,
		       price, mileage, status, description, images, metadata,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM cars
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var car Car
	err := r.db.QueryRow(query, id, r.tenantID).Scan(
		&car.ID,
		&car.TenantID,
		&car.Make,
		&car.Model,
		&car.Year,
		&car.Color,
		&car.VIN,
		&car.StockNumber,
		&car.Price,
		&car.Mileage,
		&car.Status,
		&car.Description,
		&car.Images,
		&car.Metadata,
		&car.CreatedAt,
		&car.UpdatedAt,
		&car.CreatedBy,
		&car.UpdatedBy,
		&car.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return Car{}, ErrNotFound
	}
	if err != nil {
		return Car{}, fmt.Errorf("failed to get car: %w", err)
	}

	return car, nil
}

// GetAll retrieves all cars (within tenant scope)
func (r *PostgresRepository) GetAll() []Car {
	query := `
		SELECT id, tenant_id, make, model, year, color, vin, stock_number,
		       price, mileage, status, description, images, metadata,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM cars
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, r.tenantID)
	if err != nil {
		// In production, we should log this error
		return []Car{}
	}
	defer rows.Close()

	var cars []Car
	for rows.Next() {
		var car Car
		err := rows.Scan(
			&car.ID,
			&car.TenantID,
			&car.Make,
			&car.Model,
			&car.Year,
			&car.Color,
			&car.VIN,
			&car.StockNumber,
			&car.Price,
			&car.Mileage,
			&car.Status,
			&car.Description,
			&car.Images,
			&car.Metadata,
			&car.CreatedAt,
			&car.UpdatedAt,
			&car.CreatedBy,
			&car.UpdatedBy,
			&car.DeletedAt,
		)
		if err != nil {
			// In production, we should log this error
			continue
		}
		cars = append(cars, car)
	}

	return cars
}

// Create adds a new car to the repository (within tenant scope)
func (r *PostgresRepository) Create(car Car) (Car, error) {
	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	// Set tenant ID from context
	car.TenantID = r.tenantID

	// Set default status if not provided
	if car.Status == "" {
		car.Status = "available"
	}

	// Initialize JSONB fields if nil
	if car.Images == nil {
		car.Images = JSONArray{}
	}
	if car.Metadata == nil {
		car.Metadata = JSONB{}
	}

	query := `
		INSERT INTO cars (
			id, tenant_id, make, model, year, color, vin, stock_number,
			price, mileage, status, description, images, metadata, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		car.ID,
		car.TenantID,
		car.Make,
		car.Model,
		car.Year,
		car.Color,
		car.VIN,
		car.StockNumber,
		car.Price,
		car.Mileage,
		car.Status,
		car.Description,
		car.Images,
		car.Metadata,
		car.CreatedBy,
	).Scan(&car.CreatedAt, &car.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation
		if err.Error() == `pq: duplicate key value violates unique constraint "cars_pkey"` {
			return Car{}, errors.New("car with this ID already exists")
		}
		return Car{}, fmt.Errorf("failed to create car: %w", err)
	}

	return car, nil
}

// Update updates an existing car (within tenant scope)
func (r *PostgresRepository) Update(car Car) (Car, error) {
	if car.ID == "" {
		return Car{}, ErrInvalidID
	}

	// Ensure tenant ID matches
	car.TenantID = r.tenantID

	query := `
		UPDATE cars
		SET make = $1, model = $2, year = $3, color = $4, vin = $5,
		    stock_number = $6, price = $7, mileage = $8, status = $9,
		    description = $10, images = $11, metadata = $12, updated_by = $13,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $14 AND tenant_id = $15 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		car.Make,
		car.Model,
		car.Year,
		car.Color,
		car.VIN,
		car.StockNumber,
		car.Price,
		car.Mileage,
		car.Status,
		car.Description,
		car.Images,
		car.Metadata,
		car.UpdatedBy,
		car.ID,
		car.TenantID,
	).Scan(&car.UpdatedAt)

	if err == sql.ErrNoRows {
		return Car{}, ErrNotFound
	}
	if err != nil {
		return Car{}, fmt.Errorf("failed to update car: %w", err)
	}

	return car, nil
}

// Delete soft-deletes a car from the repository (within tenant scope)
func (r *PostgresRepository) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	query := `
		UPDATE cars
		SET deleted_at = $1
		WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, time.Now(), id, r.tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete car: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetByFilters retrieves cars with filtering, sorting, and pagination (within tenant scope)
func (r *PostgresRepository) GetByFilters(filters map[string]string, sortBy string, sortOrder string, page, pageSize int) ([]Car, int, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{r.tenantID}
	argIndex := 2

	for key, value := range filters {
		if value != "" {
			switch key {
			case "make", "model", "color", "status":
				whereClause += fmt.Sprintf(" AND LOWER(%s) = LOWER($%d)", key, argIndex)
				args = append(args, value)
				argIndex++
			case "year":
				whereClause += fmt.Sprintf(" AND year = $%d", argIndex)
				args = append(args, value)
				argIndex++
			case "vin":
				whereClause += fmt.Sprintf(" AND vin = $%d", argIndex)
				args = append(args, value)
				argIndex++
			case "stock_number":
				whereClause += fmt.Sprintf(" AND stock_number = $%d", argIndex)
				args = append(args, value)
				argIndex++
			}
		}
	}

	// Count total matching records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cars %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cars: %w", err)
	}

	// Build ORDER BY clause
	orderByClause := "ORDER BY created_at DESC"
	if sortBy != "" {
		order := "ASC"
		if sortOrder == "desc" {
			order = "DESC"
		}
		orderByClause = fmt.Sprintf("ORDER BY %s %s", sortBy, order)
	}

	// Build LIMIT and OFFSET
	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf("LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Build full query
	query := fmt.Sprintf(`
		SELECT id, tenant_id, make, model, year, color, vin, stock_number,
		       price, mileage, status, description, images, metadata,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM cars
		%s
		%s
		%s
	`, whereClause, orderByClause, limitClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query cars: %w", err)
	}
	defer rows.Close()

	var cars []Car
	for rows.Next() {
		var car Car
		err := rows.Scan(
			&car.ID,
			&car.TenantID,
			&car.Make,
			&car.Model,
			&car.Year,
			&car.Color,
			&car.VIN,
			&car.StockNumber,
			&car.Price,
			&car.Mileage,
			&car.Status,
			&car.Description,
			&car.Images,
			&car.Metadata,
			&car.CreatedAt,
			&car.UpdatedAt,
			&car.CreatedBy,
			&car.UpdatedBy,
			&car.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan car: %w", err)
		}
		cars = append(cars, car)
	}

	return cars, total, nil
}
