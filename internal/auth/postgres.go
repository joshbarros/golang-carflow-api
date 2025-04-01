package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// PostgresConfig holds the configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) (*PostgresRepository, error) {
	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// CreateUser creates a new user in the database
func (r *PostgresRepository) CreateUser(user User) (User, error) {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name,
			role, tenant_id, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING *`

	err := r.db.Get(&user, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Role, user.TenantID, user.Active, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(id string) (User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(email string) (User, error) {
	var user User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.Get(&user, query, email)
	if err != nil {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates an existing user
func (r *PostgresRepository) UpdateUser(user User) (User, error) {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			email = $1,
			password_hash = $2,
			first_name = $3,
			last_name = $4,
			role = $5,
			tenant_id = $6,
			active = $7,
			updated_at = $8
		WHERE id = $9
		RETURNING *`

	err := r.db.Get(&user, query,
		user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Role, user.TenantID, user.Active, user.UpdatedAt, user.ID)

	if err != nil {
		return User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *PostgresRepository) UpdateLastLogin(id string) error {
	query := `
		UPDATE users SET
			last_login_at = $1,
			updated_at = $1
		WHERE id = $2`

	now := time.Now()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser removes a user from the database
func (r *PostgresRepository) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsersByTenant retrieves all users for a specific tenant
func (r *PostgresRepository) ListUsersByTenant(tenantID string) ([]User, error) {
	var users []User
	query := `SELECT * FROM users WHERE tenant_id = $1`
	err := r.db.Select(&users, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
