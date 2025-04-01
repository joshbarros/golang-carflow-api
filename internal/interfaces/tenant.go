package interfaces

import "github.com/joshbarros/golang-carflow-api/internal/domain"

// TenantService defines the interface for tenant operations
type TenantService interface {
	CreateTenant(tenant domain.Tenant) error
	GetTenant(id string) (*domain.Tenant, error)
	UpdateTenant(tenant domain.Tenant) error
	DeleteTenant(id string) error
	ListTenants(page, pageSize int) ([]domain.Tenant, error)
	GetTenantByDomain(domain string) (*domain.Tenant, error)
}
