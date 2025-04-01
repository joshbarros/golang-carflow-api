package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Plan types
const (
	PlanBasic      = "basic"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// Tenant statuses
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
)

// Feature represents a feature flag for a tenant
type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// ResourceLimits represents the resource limits for a tenant
type ResourceLimits struct {
	MaxUsers        int `json:"max_users"`
	MaxCars         int `json:"max_cars"`
	APIRateLimit    int `json:"api_rate_limit"`
	StorageLimit    int `json:"storage_limit"`
	BackupRetention int `json:"backup_retention"`
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Email                string         `json:"email"`
	Plan                 string         `json:"plan"`
	Features             []Feature      `json:"features"`
	Limits               ResourceLimits `json:"limits"`
	CustomDomain         string         `json:"custom_domain"`
	Status               string         `json:"status"`
	StripeCustomerID     string         `json:"stripe_customer_id"`
	StripeSubscriptionID string         `json:"stripe_subscription_id"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// Error represents a domain error
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new domain error
func NewError(message string) error {
	return &Error{Message: message}
}

// Common errors
var (
	ErrTenantNotFound = NewError("tenant not found")
	ErrTenantExists   = NewError("tenant already exists")
)

// GetDefaultResourceLimits returns the default resource limits for a plan
func GetDefaultResourceLimits(plan string) ResourceLimits {
	switch plan {
	case PlanBasic:
		return ResourceLimits{
			MaxUsers:        10,
			MaxCars:         50,
			APIRateLimit:    100,
			StorageLimit:    1024, // 1GB in MB
			BackupRetention: 7,    // 7 days
		}
	case PlanPro:
		return ResourceLimits{
			MaxUsers:        50,
			MaxCars:         200,
			APIRateLimit:    500,
			StorageLimit:    5120, // 5GB in MB
			BackupRetention: 30,   // 30 days
		}
	case PlanEnterprise:
		return ResourceLimits{
			MaxUsers:        -1, // Unlimited
			MaxCars:         -1, // Unlimited
			APIRateLimit:    1000,
			StorageLimit:    -1, // Unlimited
			BackupRetention: 90, // 90 days
		}
	default:
		return ResourceLimits{}
	}
}

// GetDefaultFeatures returns the default features for a plan
func GetDefaultFeatures(plan string) []Feature {
	basic := []Feature{
		{Name: "basic_reporting", Description: "Basic reporting and analytics", Enabled: true},
		{Name: "email_support", Description: "Email support during business hours", Enabled: true},
	}

	pro := append(basic,
		Feature{Name: "advanced_reporting", Description: "Advanced reporting and analytics", Enabled: true},
		Feature{Name: "priority_support", Description: "Priority email and phone support", Enabled: true},
		Feature{Name: "api_access", Description: "API access for integration", Enabled: true},
	)

	enterprise := append(pro,
		Feature{Name: "custom_branding", Description: "Custom branding and white-labeling", Enabled: true},
		Feature{Name: "dedicated_support", Description: "24/7 dedicated support", Enabled: true},
		Feature{Name: "audit_logs", Description: "Detailed audit logs", Enabled: true},
	)

	switch plan {
	case PlanBasic:
		return basic
	case PlanPro:
		return pro
	case PlanEnterprise:
		return enterprise
	default:
		return []Feature{}
	}
}

// IsValidPlan checks if a plan is valid
func IsValidPlan(plan string) bool {
	switch plan {
	case PlanBasic, PlanPro, PlanEnterprise:
		return true
	default:
		return false
	}
}

// IsValidStatus checks if a status is valid
func IsValidStatus(status string) bool {
	switch status {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}

// ValidateDomain checks if a domain is valid
func ValidateDomain(domain string) error {
	// Basic domain format validation
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format")
	}

	// Parse domain to check for valid URL
	u, err := url.Parse("https://" + domain)
	if err != nil {
		return fmt.Errorf("invalid domain: %v", err)
	}

	// Check for common TLDs
	parts := strings.Split(u.Hostname(), ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least one subdomain")
	}

	return nil
}
