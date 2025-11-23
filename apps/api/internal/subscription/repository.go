package subscription

import (
	"database/sql"
	"fmt"
	"time"
)

// Subscription represents a subscription in the system
type Subscription struct {
	ID                   string     `json:"id" db:"id"`
	TenantID             string     `json:"tenant_id" db:"tenant_id"`
	StripeCustomerID     string     `json:"stripe_customer_id" db:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	StripePriceID        string     `json:"stripe_price_id" db:"stripe_price_id"`
	Plan                 string     `json:"plan" db:"plan"`
	Status               string     `json:"status" db:"status"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// Status constants
const (
	StatusTrialing = "trialing"
	StatusActive   = "active"
	StatusPastDue  = "past_due"
	StatusCanceled = "canceled"
	StatusUnpaid   = "unpaid"
)

// Repository defines the interface for subscription data access
type Repository interface {
	Create(subscription *Subscription) error
	GetByID(id string) (*Subscription, error)
	GetByTenantID(tenantID string) (*Subscription, error)
	GetByStripeCustomerID(customerID string) (*Subscription, error)
	GetByStripeSubscriptionID(subscriptionID string) (*Subscription, error)
	Update(subscription *Subscription) error
	Delete(id string) error
}

// PostgresRepository implements Repository interface with PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL subscription repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create adds a new subscription to the database
func (r *PostgresRepository) Create(subscription *Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, tenant_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
			plan, status, current_period_start, current_period_end,
			cancel_at_period_end, canceled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		subscription.ID,
		subscription.TenantID,
		subscription.StripeCustomerID,
		subscription.StripeSubscriptionID,
		subscription.StripePriceID,
		subscription.Plan,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
	).Scan(&subscription.CreatedAt, &subscription.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by ID
func (r *PostgresRepository) GetByID(id string) (*Subscription, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       plan, status, current_period_start, current_period_end,
		       cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	subscription := &Subscription{}
	err := r.db.QueryRow(query, id).Scan(
		&subscription.ID,
		&subscription.TenantID,
		&subscription.StripeCustomerID,
		&subscription.StripeSubscriptionID,
		&subscription.StripePriceID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CanceledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// GetByTenantID retrieves a subscription by tenant ID
func (r *PostgresRepository) GetByTenantID(tenantID string) (*Subscription, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       plan, status, current_period_start, current_period_end,
		       cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	subscription := &Subscription{}
	err := r.db.QueryRow(query, tenantID).Scan(
		&subscription.ID,
		&subscription.TenantID,
		&subscription.StripeCustomerID,
		&subscription.StripeSubscriptionID,
		&subscription.StripePriceID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CanceledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No subscription found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// GetByStripeCustomerID retrieves a subscription by Stripe customer ID
func (r *PostgresRepository) GetByStripeCustomerID(customerID string) (*Subscription, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       plan, status, current_period_start, current_period_end,
		       cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_customer_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	subscription := &Subscription{}
	err := r.db.QueryRow(query, customerID).Scan(
		&subscription.ID,
		&subscription.TenantID,
		&subscription.StripeCustomerID,
		&subscription.StripeSubscriptionID,
		&subscription.StripePriceID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CanceledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// GetByStripeSubscriptionID retrieves a subscription by Stripe subscription ID
func (r *PostgresRepository) GetByStripeSubscriptionID(subscriptionID string) (*Subscription, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       plan, status, current_period_start, current_period_end,
		       cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	subscription := &Subscription{}
	err := r.db.QueryRow(query, subscriptionID).Scan(
		&subscription.ID,
		&subscription.TenantID,
		&subscription.StripeCustomerID,
		&subscription.StripeSubscriptionID,
		&subscription.StripePriceID,
		&subscription.Plan,
		&subscription.Status,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&subscription.CanceledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// Update updates an existing subscription
func (r *PostgresRepository) Update(subscription *Subscription) error {
	query := `
		UPDATE subscriptions
		SET stripe_customer_id = $1, stripe_subscription_id = $2, stripe_price_id = $3,
		    plan = $4, status = $5, current_period_start = $6, current_period_end = $7,
		    cancel_at_period_end = $8, canceled_at = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		subscription.StripeCustomerID,
		subscription.StripeSubscriptionID,
		subscription.StripePriceID,
		subscription.Plan,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
		subscription.ID,
	).Scan(&subscription.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("subscription not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// Delete removes a subscription
func (r *PostgresRepository) Delete(id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}
