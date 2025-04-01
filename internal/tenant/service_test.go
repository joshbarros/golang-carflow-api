package tenant

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshbarros/golang-carflow-api/internal/domain"
)

// mockRepository is a mock implementation of the Repository interface
type mockRepository struct {
	tenants map[string]*domain.Tenant
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		tenants: make(map[string]*domain.Tenant),
	}
}

func (r *mockRepository) Create(tenant domain.Tenant) error {
	if _, exists := r.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant already exists")
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	r.tenants[tenant.ID] = &tenant
	return nil
}

func (r *mockRepository) Get(id string) (*domain.Tenant, error) {
	if tenant, exists := r.tenants[id]; exists {
		return tenant, nil
	}
	return nil, fmt.Errorf("tenant not found: %s", id)
}

func (r *mockRepository) Update(tenant domain.Tenant) error {
	if _, exists := r.tenants[tenant.ID]; !exists {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}
	tenant.UpdatedAt = time.Now()
	r.tenants[tenant.ID] = &tenant
	return nil
}

func (r *mockRepository) Delete(id string) error {
	if _, exists := r.tenants[id]; !exists {
		return fmt.Errorf("tenant not found: %s", id)
	}
	delete(r.tenants, id)
	return nil
}

func (r *mockRepository) List(page, pageSize int) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	for _, t := range r.tenants {
		tenants = append(tenants, *t)
	}
	return tenants, nil
}

func (r *mockRepository) GetByDomain(domain string) (*domain.Tenant, error) {
	for _, t := range r.tenants {
		if t.CustomDomain == domain {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tenant not found for domain: %s", domain)
}

func TestCreateTenant(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)

	tests := []struct {
		name    string
		tenant  domain.Tenant
		wantErr bool
	}{
		{
			name: "Valid tenant",
			tenant: domain.Tenant{
				Name:         "Test Tenant",
				Plan:         domain.PlanBasic,
				CustomDomain: "test.example.com",
			},
			wantErr: false,
		},
		{
			name: "Invalid name - too short",
			tenant: domain.Tenant{
				Name:         "Te",
				Plan:         domain.PlanBasic,
				CustomDomain: "test.example.com",
			},
			wantErr: true,
		},
		{
			name: "Invalid plan",
			tenant: domain.Tenant{
				Name:         "Test Tenant",
				Plan:         "invalid",
				CustomDomain: "test.example.com",
			},
			wantErr: true,
		},
		{
			name: "Invalid domain",
			tenant: domain.Tenant{
				Name:         "Test Tenant",
				Plan:         domain.PlanBasic,
				CustomDomain: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateTenant(tt.tenant)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateTenant(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)

	// Create initial tenant
	tenant := domain.Tenant{
		ID:           uuid.New().String(),
		Name:         "Test Tenant",
		Plan:         domain.PlanBasic,
		CustomDomain: "test.example.com",
		Status:       domain.StatusActive,
	}
	if err := service.CreateTenant(tenant); err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	tests := []struct {
		name    string
		tenant  domain.Tenant
		wantErr bool
	}{
		{
			name: "Valid update",
			tenant: domain.Tenant{
				ID:           tenant.ID,
				Name:         "Updated Tenant",
				Plan:         domain.PlanPro,
				CustomDomain: "updated.example.com",
				Status:       domain.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "Non-existent tenant",
			tenant: domain.Tenant{
				ID:           uuid.New().String(),
				Name:         "Non-existent",
				Plan:         domain.PlanBasic,
				CustomDomain: "test.example.com",
				Status:       domain.StatusActive,
			},
			wantErr: true,
		},
		{
			name: "Invalid name",
			tenant: domain.Tenant{
				ID:           tenant.ID,
				Name:         "Te",
				Plan:         domain.PlanBasic,
				CustomDomain: "test.example.com",
				Status:       domain.StatusActive,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateTenant(tt.tenant)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTenant() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify the update
				updated, err := service.GetTenant(tt.tenant.ID)
				if err != nil {
					t.Errorf("Failed to get updated tenant: %v", err)
				} else {
					if updated.Name != tt.tenant.Name {
						t.Errorf("Name not updated: got %v, want %v", updated.Name, tt.tenant.Name)
					}
					if updated.Plan != tt.tenant.Plan {
						t.Errorf("Plan not updated: got %v, want %v", updated.Plan, tt.tenant.Plan)
					}
					if updated.CustomDomain != tt.tenant.CustomDomain {
						t.Errorf("CustomDomain not updated: got %v, want %v", updated.CustomDomain, tt.tenant.CustomDomain)
					}
				}
			}
		})
	}
}

func TestDeleteTenant(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)

	// Create test tenant
	tenant := domain.Tenant{
		ID:           uuid.New().String(),
		Name:         "Test Tenant",
		Plan:         domain.PlanBasic,
		CustomDomain: "test.example.com",
		Status:       domain.StatusActive,
	}
	if err := service.CreateTenant(tenant); err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid delete",
			id:      tenant.ID,
			wantErr: false,
		},
		{
			name:    "Non-existent tenant",
			id:      uuid.New().String(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteTenant(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTenant() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify the deletion
				_, err := service.GetTenant(tt.id)
				if err == nil {
					t.Error("Tenant still exists after deletion")
				}
			}
		})
	}
}

func TestListTenants(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo)

	// Create test tenants
	tenants := []domain.Tenant{
		{
			ID:           uuid.New().String(),
			Name:         "Tenant 1",
			Plan:         domain.PlanBasic,
			CustomDomain: "tenant1.example.com",
			Status:       domain.StatusActive,
		},
		{
			ID:           uuid.New().String(),
			Name:         "Tenant 2",
			Plan:         domain.PlanPro,
			CustomDomain: "tenant2.example.com",
			Status:       domain.StatusActive,
		},
	}

	for _, tenant := range tenants {
		if err := service.CreateTenant(tenant); err != nil {
			t.Fatalf("Failed to create test tenant: %v", err)
		}
	}

	tests := []struct {
		name      string
		page      int
		pageSize  int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "List all tenants",
			page:      1,
			pageSize:  10,
			wantCount: len(tenants),
			wantErr:   false,
		},
		{
			name:      "Empty page",
			page:      100,
			pageSize:  10,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ListTenants(tt.page, tt.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListTenants() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("ListTenants() returned %d tenants, want %d", len(got), tt.wantCount)
			}
		})
	}
}
