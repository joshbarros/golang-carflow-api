package tenant

import (
	"time"
)

// Tenant represents a dealership/company in the system
type Tenant struct {
	ID                     string     `json:"id" db:"id"`
	Name                   string     `json:"name" db:"name"`
	Slug                   string     `json:"slug" db:"slug"`
	SubscriptionStatus     string     `json:"subscription_status" db:"subscription_status"`
	SubscriptionPlan       string     `json:"subscription_plan" db:"subscription_plan"`
	TrialEndsAt            *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	MaxVehicles            int        `json:"max_vehicles" db:"max_vehicles"`
	MaxUsers               int        `json:"max_users" db:"max_users"`
	MaxAPICallsPerMonth    int        `json:"max_api_calls_per_month" db:"max_api_calls_per_month"`
	OwnerEmail             string     `json:"owner_email" db:"owner_email"`
	OwnerPhone             string     `json:"owner_phone,omitempty" db:"owner_phone"`
	CompanyAddress         string     `json:"company_address,omitempty" db:"company_address"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SubscriptionPlan constants
const (
	PlanStarter      = "starter"
	PlanProfessional = "professional"
	PlanEnterprise   = "enterprise"
)

// SubscriptionStatus constants
const (
	StatusTrial    = "trial"
	StatusActive   = "active"
	StatusPastDue  = "past_due"
	StatusCanceled = "canceled"
)

// PlanLimits returns the limits for a given plan
func PlanLimits(plan string) (maxVehicles, maxUsers, maxAPICalls int) {
	switch plan {
	case PlanStarter:
		return 100, 3, 10000
	case PlanProfessional:
		return 1000, 10, 100000
	case PlanEnterprise:
		return 999999, 999999, 999999999 // Unlimited
	default:
		return 100, 3, 10000 // Default to starter
	}
}
