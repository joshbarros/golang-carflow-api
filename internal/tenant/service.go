package tenant

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

var (
	// ErrInvalidTenant is returned when tenant data is invalid
	ErrInvalidTenant = fmt.Errorf("invalid tenant data")
	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = fmt.Errorf("tenant not found")
	// ErrDomainTaken is returned when a custom domain is already in use
	ErrDomainTaken = fmt.Errorf("custom domain is already in use")
)

// Service handles tenant business logic
type Service struct {
	repo Repository
}

// NewService creates a new tenant service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(tenant domain.Tenant) error {
	// Validate tenant data
	if err := validateTenant(tenant); err != nil {
		return err
	}

	// Generate UUID if not provided
	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}

	// Set default status if not provided
	if tenant.Status == "" {
		tenant.Status = domain.StatusActive
	}

	// Set default resource limits based on plan
	if tenant.Limits == (domain.ResourceLimits{}) {
		tenant.Limits = domain.GetDefaultResourceLimits(tenant.Plan)
	}

	// Set default features based on plan
	if len(tenant.Features) == 0 {
		tenant.Features = domain.GetDefaultFeatures(tenant.Plan)
	}

	// Check if domain is already taken
	if tenant.CustomDomain != "" {
		existing, err := s.repo.GetByDomain(tenant.CustomDomain)
		if err == nil && existing != nil {
			return ErrDomainTaken
		}
	}

	// Set timestamps
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	return s.repo.Create(tenant)
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(id string) (*domain.Tenant, error) {
	return s.repo.Get(id)
}

// UpdateTenant updates an existing tenant
func (s *Service) UpdateTenant(tenant domain.Tenant) error {
	// Validate tenant data
	if err := validateTenant(tenant); err != nil {
		return err
	}

	// Get existing tenant
	existing, err := s.repo.Get(tenant.ID)
	if err != nil {
		return err
	}

	// Check if domain is already taken by another tenant
	if tenant.CustomDomain != "" && tenant.CustomDomain != existing.CustomDomain {
		t, err := s.repo.GetByDomain(tenant.CustomDomain)
		if err == nil && t != nil {
			return ErrDomainTaken
		}
	}

	// Update resource limits if plan changed
	if tenant.Plan != existing.Plan {
		tenant.Limits = domain.GetDefaultResourceLimits(tenant.Plan)
		tenant.Features = domain.GetDefaultFeatures(tenant.Plan)
	}

	// Update timestamp
	tenant.UpdatedAt = time.Now()

	return s.repo.Update(tenant)
}

// DeleteTenant deletes a tenant by ID
func (s *Service) DeleteTenant(id string) error {
	// Check if tenant exists
	if _, err := s.repo.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

// ListTenants retrieves a paginated list of tenants
func (s *Service) ListTenants(page, pageSize int) ([]domain.Tenant, error) {
	// Set default values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return s.repo.List(page, pageSize)
}

// GetTenantByDomain retrieves a tenant by custom domain
func (s *Service) GetTenantByDomain(domain string) (*domain.Tenant, error) {
	return s.repo.GetByDomain(domain)
}

// validateTenant validates tenant data
func validateTenant(tenant domain.Tenant) error {
	// Validate name
	if len(tenant.Name) < 3 {
		return fmt.Errorf("tenant name must be at least 3 characters long")
	}

	// Validate plan
	if !domain.IsValidPlan(tenant.Plan) {
		return fmt.Errorf("invalid plan: %s", tenant.Plan)
	}

	// Validate custom domain if provided
	if tenant.CustomDomain != "" {
		if err := domain.ValidateDomain(tenant.CustomDomain); err != nil {
			return fmt.Errorf("invalid custom domain: %v", err)
		}
	}

	// Validate status if provided
	if tenant.Status != "" && !domain.IsValidStatus(tenant.Status) {
		return fmt.Errorf("invalid status: %s", tenant.Status)
	}

	return nil
}

// validateDomain checks if a domain is valid
func validateDomain(domain string) error {
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
