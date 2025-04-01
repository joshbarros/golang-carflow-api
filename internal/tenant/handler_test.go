package tenant

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// mockService is a mock implementation of the tenant service
type mockService struct {
	tenants map[string]*Tenant
}

func newMockService() *mockService {
	return &mockService{
		tenants: make(map[string]*Tenant),
	}
}

func (s *mockService) CreateTenant(tenant Tenant) error {
	if _, exists := s.tenants[tenant.ID]; exists {
		return ErrTenantExists
	}
	s.tenants[tenant.ID] = &tenant
	return nil
}

func (s *mockService) GetTenant(id string) (*Tenant, error) {
	if tenant, exists := s.tenants[id]; exists {
		return tenant, nil
	}
	return nil, ErrTenantNotFound
}

func (s *mockService) UpdateTenant(tenant Tenant) error {
	if _, exists := s.tenants[tenant.ID]; !exists {
		return ErrTenantNotFound
	}
	s.tenants[tenant.ID] = &tenant
	return nil
}

func (s *mockService) DeleteTenant(id string) error {
	if _, exists := s.tenants[id]; !exists {
		return ErrTenantNotFound
	}
	delete(s.tenants, id)
	return nil
}

func (s *mockService) ListTenants(page, pageSize int) ([]Tenant, error) {
	var tenants []Tenant
	for _, t := range s.tenants {
		tenants = append(tenants, *t)
	}
	return tenants, nil
}

func (s *mockService) GetTenantByDomain(domain string) (*Tenant, error) {
	for _, t := range s.tenants {
		if t.CustomDomain == domain {
			return t, nil
		}
	}
	return nil, ErrTenantNotFound
}

func TestHandleListTenants(t *testing.T) {
	service := newMockService()
	handler := NewHandler(service)

	// Create test tenants
	tenants := []Tenant{
		{
			ID:           uuid.New().String(),
			Name:         "Tenant 1",
			Plan:         PlanBasic,
			CustomDomain: "tenant1.example.com",
			Status:       StatusActive,
		},
		{
			ID:           uuid.New().String(),
			Name:         "Tenant 2",
			Plan:         PlanPro,
			CustomDomain: "tenant2.example.com",
			Status:       StatusActive,
		},
	}

	for _, tenant := range tenants {
		if err := service.CreateTenant(tenant); err != nil {
			t.Fatalf("Failed to create test tenant: %v", err)
		}
	}

	tests := []struct {
		name       string
		page       string
		pageSize   string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "List all tenants",
			page:       "1",
			pageSize:   "10",
			wantStatus: http.StatusOK,
			wantCount:  len(tenants),
		},
		{
			name:       "Invalid page",
			page:       "invalid",
			pageSize:   "10",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid page size",
			page:       "1",
			pageSize:   "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tenants?page="+tt.page+"&pageSize="+tt.pageSize, nil)
			w := httptest.NewRecorder()

			handler.HandleListTenants(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("HandleListTenants() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var result []Tenant
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(result) != tt.wantCount {
					t.Errorf("HandleListTenants() returned %d tenants, want %d", len(result), tt.wantCount)
				}
			}
		})
	}
}

func TestHandleCreateTenant(t *testing.T) {
	service := newMockService()
	handler := NewHandler(service)

	tests := []struct {
		name       string
		tenant     Tenant
		wantStatus int
	}{
		{
			name: "Valid tenant",
			tenant: Tenant{
				Name:         "Test Tenant",
				Plan:         PlanBasic,
				CustomDomain: "test.example.com",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid tenant - missing name",
			tenant: Tenant{
				Plan:         PlanBasic,
				CustomDomain: "test.example.com",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid tenant - invalid plan",
			tenant: Tenant{
				Name:         "Test Tenant",
				Plan:         "invalid",
				CustomDomain: "test.example.com",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.tenant)
			if err != nil {
				t.Fatalf("Failed to marshal tenant: %v", err)
			}

			req := httptest.NewRequest("POST", "/tenants", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleCreateTenant(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("HandleCreateTenant() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetTenant(t *testing.T) {
	service := newMockService()
	handler := NewHandler(service)

	// Create test tenant
	tenant := Tenant{
		ID:           uuid.New().String(),
		Name:         "Test Tenant",
		Plan:         PlanBasic,
		CustomDomain: "test.example.com",
		Status:       StatusActive,
	}
	if err := service.CreateTenant(tenant); err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	tests := []struct {
		name       string
		tenantID   string
		wantStatus int
	}{
		{
			name:       "Valid tenant",
			tenantID:   tenant.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Non-existent tenant",
			tenantID:   uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Invalid tenant ID",
			tenantID:   "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/tenants/"+tt.tenantID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.tenantID)
			req = req.WithContext(chi.NewContext(req.Context(), rctx))

			w := httptest.NewRecorder()

			handler.HandleGetTenant(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("HandleGetTenant() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var result Tenant
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if result.ID != tt.tenantID {
					t.Errorf("HandleGetTenant() returned tenant ID = %v, want %v", result.ID, tt.tenantID)
				}
			}
		})
	}
}

func TestHandleUpdateTenant(t *testing.T) {
	service := newMockService()
	handler := NewHandler(service)

	// Create test tenant
	tenant := Tenant{
		ID:           uuid.New().String(),
		Name:         "Test Tenant",
		Plan:         PlanBasic,
		CustomDomain: "test.example.com",
		Status:       StatusActive,
	}
	if err := service.CreateTenant(tenant); err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	tests := []struct {
		name       string
		tenantID   string
		tenant     Tenant
		wantStatus int
	}{
		{
			name:     "Valid update",
			tenantID: tenant.ID,
			tenant: Tenant{
				ID:           tenant.ID,
				Name:         "Updated Tenant",
				Plan:         PlanPro,
				CustomDomain: "updated.example.com",
				Status:       StatusActive,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "Non-existent tenant",
			tenantID: uuid.New().String(),
			tenant: Tenant{
				ID:           uuid.New().String(),
				Name:         "Non-existent",
				Plan:         PlanBasic,
				CustomDomain: "test.example.com",
				Status:       StatusActive,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "Invalid tenant data",
			tenantID: tenant.ID,
			tenant: Tenant{
				ID:           tenant.ID,
				Name:         "",
				Plan:         "invalid",
				CustomDomain: "test.example.com",
				Status:       StatusActive,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.tenant)
			if err != nil {
				t.Fatalf("Failed to marshal tenant: %v", err)
			}

			req := httptest.NewRequest("PUT", "/tenants/"+tt.tenantID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.tenantID)
			req = req.WithContext(chi.NewContext(req.Context(), rctx))

			w := httptest.NewRecorder()

			handler.HandleUpdateTenant(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("HandleUpdateTenant() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHandleDeleteTenant(t *testing.T) {
	service := newMockService()
	handler := NewHandler(service)

	// Create test tenant
	tenant := Tenant{
		ID:           uuid.New().String(),
		Name:         "Test Tenant",
		Plan:         PlanBasic,
		CustomDomain: "test.example.com",
		Status:       StatusActive,
	}
	if err := service.CreateTenant(tenant); err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	tests := []struct {
		name       string
		tenantID   string
		wantStatus int
	}{
		{
			name:       "Valid delete",
			tenantID:   tenant.ID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Non-existent tenant",
			tenantID:   uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Invalid tenant ID",
			tenantID:   "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/tenants/"+tt.tenantID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.tenantID)
			req = req.WithContext(chi.NewContext(req.Context(), rctx))

			w := httptest.NewRecorder()

			handler.HandleDeleteTenant(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("HandleDeleteTenant() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
