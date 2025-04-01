package tenant

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

// Repository defines the interface for tenant data storage
type Repository interface {
	Create(tenant domain.Tenant) error
	Get(id string) (*domain.Tenant, error)
	GetByDomain(domain string) (*domain.Tenant, error)
	Update(tenant domain.Tenant) error
	Delete(id string) error
	List(page, pageSize int) ([]domain.Tenant, error)
}

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	return &PostgresRepository{db: db}, nil
}

// Create inserts a new tenant into the database
func (r *PostgresRepository) Create(tenant domain.Tenant) error {
	// Check if tenant with same domain exists
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE custom_domain = $1)", tenant.CustomDomain).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking domain existence: %v", err)
	}
	if exists {
		return domain.ErrTenantExists
	}

	// Marshal features and limits to JSON
	features, err := json.Marshal(tenant.Features)
	if err != nil {
		return fmt.Errorf("error marshaling features: %v", err)
	}

	limits, err := json.Marshal(tenant.Limits)
	if err != nil {
		return fmt.Errorf("error marshaling limits: %v", err)
	}

	// Insert tenant
	query := `
		INSERT INTO tenants (id, name, plan, features, limits, custom_domain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = r.db.Exec(query,
		tenant.ID,
		tenant.Name,
		tenant.Plan,
		features,
		limits,
		tenant.CustomDomain,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating tenant: %v", err)
	}

	return nil
}

// Get retrieves a tenant by ID
func (r *PostgresRepository) Get(id string) (*domain.Tenant, error) {
	tenant := &domain.Tenant{}
	var features, limits []byte

	query := `
		SELECT id, name, plan, features, limits, custom_domain, status, created_at, updated_at
		FROM tenants
		WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Plan,
		&features,
		&limits,
		&tenant.CustomDomain,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tenant: %v", err)
	}

	// Unmarshal features and limits
	if err := json.Unmarshal(features, &tenant.Features); err != nil {
		return nil, fmt.Errorf("error unmarshaling features: %v", err)
	}
	if err := json.Unmarshal(limits, &tenant.Limits); err != nil {
		return nil, fmt.Errorf("error unmarshaling limits: %v", err)
	}

	return tenant, nil
}

// Update modifies an existing tenant
func (r *PostgresRepository) Update(tenant domain.Tenant) error {
	// Check if tenant exists
	_, err := r.Get(tenant.ID)
	if err != nil {
		return err
	}

	// Marshal features and limits to JSON
	features, err := json.Marshal(tenant.Features)
	if err != nil {
		return fmt.Errorf("error marshaling features: %v", err)
	}

	limits, err := json.Marshal(tenant.Limits)
	if err != nil {
		return fmt.Errorf("error marshaling limits: %v", err)
	}

	// Update tenant
	query := `
		UPDATE tenants
		SET name = $1, plan = $2, features = $3, limits = $4, custom_domain = $5, status = $6, updated_at = NOW()
		WHERE id = $7`
	result, err := r.db.Exec(query,
		tenant.Name,
		tenant.Plan,
		features,
		limits,
		tenant.CustomDomain,
		tenant.Status,
		tenant.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating tenant: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rows == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// Delete removes a tenant from the database
func (r *PostgresRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM tenants WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting tenant: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rows == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// List returns a paginated list of tenants
func (r *PostgresRepository) List(page, pageSize int) ([]domain.Tenant, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, plan, features, limits, custom_domain, status, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing tenants: %v", err)
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var tenant domain.Tenant
		var features, limits []byte

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Plan,
			&features,
			&limits,
			&tenant.CustomDomain,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tenant row: %v", err)
		}

		// Unmarshal features and limits
		if err := json.Unmarshal(features, &tenant.Features); err != nil {
			return nil, fmt.Errorf("error unmarshaling features: %v", err)
		}
		if err := json.Unmarshal(limits, &tenant.Limits); err != nil {
			return nil, fmt.Errorf("error unmarshaling limits: %v", err)
		}

		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenant rows: %v", err)
	}

	return tenants, nil
}

// GetByDomain retrieves a tenant by custom domain
func (r *PostgresRepository) GetByDomain(domain string) (*domain.Tenant, error) {
	tenant := &domain.Tenant{}
	var features, limits []byte

	query := `
		SELECT id, name, plan, features, limits, custom_domain, status, created_at, updated_at
		FROM tenants
		WHERE custom_domain = $1`
	err := r.db.QueryRow(query, domain).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Plan,
		&features,
		&limits,
		&tenant.CustomDomain,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tenant by domain: %v", err)
	}

	// Unmarshal features and limits
	if err := json.Unmarshal(features, &tenant.Features); err != nil {
		return nil, fmt.Errorf("error unmarshaling features: %v", err)
	}
	if err := json.Unmarshal(limits, &tenant.Limits); err != nil {
		return nil, fmt.Errorf("error unmarshaling limits: %v", err)
	}

	return tenant, nil
}
