package tenant

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository defines the interface for tenant data access
type Repository interface {
	Create(tenant *Tenant) error
	GetByID(id string) (*Tenant, error)
	GetBySlug(slug string) (*Tenant, error)
	GetByEmail(email string) (*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id string) error
}

// PostgresRepository implements Repository interface with PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL tenant repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create adds a new tenant to the database
func (r *PostgresRepository) Create(tenant *Tenant) error {
	query := `
		INSERT INTO tenants (
			id, name, slug, subscription_status, subscription_plan, trial_ends_at,
			max_vehicles, max_users, max_api_calls_per_month, owner_email,
			owner_phone, company_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.SubscriptionStatus,
		tenant.SubscriptionPlan,
		tenant.TrialEndsAt,
		tenant.MaxVehicles,
		tenant.MaxUsers,
		tenant.MaxAPICallsPerMonth,
		tenant.OwnerEmail,
		tenant.OwnerPhone,
		tenant.CompanyAddress,
	).Scan(&tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID retrieves a tenant by ID
func (r *PostgresRepository) GetByID(id string) (*Tenant, error) {
	query := `
		SELECT id, name, slug, subscription_status, subscription_plan, trial_ends_at,
		       max_vehicles, max_users, max_api_calls_per_month, owner_email,
		       owner_phone, company_address, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`

	tenant := &Tenant{}
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.SubscriptionStatus,
		&tenant.SubscriptionPlan,
		&tenant.TrialEndsAt,
		&tenant.MaxVehicles,
		&tenant.MaxUsers,
		&tenant.MaxAPICallsPerMonth,
		&tenant.OwnerEmail,
		&tenant.OwnerPhone,
		&tenant.CompanyAddress,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *PostgresRepository) GetBySlug(slug string) (*Tenant, error) {
	query := `
		SELECT id, name, slug, subscription_status, subscription_plan, trial_ends_at,
		       max_vehicles, max_users, max_api_calls_per_month, owner_email,
		       owner_phone, company_address, created_at, updated_at, deleted_at
		FROM tenants
		WHERE slug = $1 AND deleted_at IS NULL
	`

	tenant := &Tenant{}
	err := r.db.QueryRow(query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.SubscriptionStatus,
		&tenant.SubscriptionPlan,
		&tenant.TrialEndsAt,
		&tenant.MaxVehicles,
		&tenant.MaxUsers,
		&tenant.MaxAPICallsPerMonth,
		&tenant.OwnerEmail,
		&tenant.OwnerPhone,
		&tenant.CompanyAddress,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetByEmail retrieves a tenant by owner email
func (r *PostgresRepository) GetByEmail(email string) (*Tenant, error) {
	query := `
		SELECT id, name, slug, subscription_status, subscription_plan, trial_ends_at,
		       max_vehicles, max_users, max_api_calls_per_month, owner_email,
		       owner_phone, company_address, created_at, updated_at, deleted_at
		FROM tenants
		WHERE owner_email = $1 AND deleted_at IS NULL
	`

	tenant := &Tenant{}
	err := r.db.QueryRow(query, email).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.SubscriptionStatus,
		&tenant.SubscriptionPlan,
		&tenant.TrialEndsAt,
		&tenant.MaxVehicles,
		&tenant.MaxUsers,
		&tenant.MaxAPICallsPerMonth,
		&tenant.OwnerEmail,
		&tenant.OwnerPhone,
		&tenant.CompanyAddress,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Email not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// Update updates an existing tenant
func (r *PostgresRepository) Update(tenant *Tenant) error {
	query := `
		UPDATE tenants
		SET name = $1, subscription_status = $2, subscription_plan = $3,
		    trial_ends_at = $4, max_vehicles = $5, max_users = $6,
		    max_api_calls_per_month = $7, owner_email = $8, owner_phone = $9,
		    company_address = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		tenant.Name,
		tenant.SubscriptionStatus,
		tenant.SubscriptionPlan,
		tenant.TrialEndsAt,
		tenant.MaxVehicles,
		tenant.MaxUsers,
		tenant.MaxAPICallsPerMonth,
		tenant.OwnerEmail,
		tenant.OwnerPhone,
		tenant.CompanyAddress,
		tenant.ID,
	).Scan(&tenant.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("tenant not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}

// Delete soft-deletes a tenant
func (r *PostgresRepository) Delete(id string) error {
	query := `
		UPDATE tenants
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}
