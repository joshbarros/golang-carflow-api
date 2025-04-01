package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v74"
)

// MockTenantService is a mock implementation of the domain.TenantService interface
type MockTenantService struct {
	tenants map[string]*domain.Tenant
}

func NewMockTenantService() *MockTenantService {
	return &MockTenantService{
		tenants: make(map[string]*domain.Tenant),
	}
}

func (m *MockTenantService) CreateTenant(tenant domain.Tenant) error {
	m.tenants[tenant.ID] = &tenant
	return nil
}

func (m *MockTenantService) UpdateTenant(tenant domain.Tenant) error {
	if _, exists := m.tenants[tenant.ID]; !exists {
		return fmt.Errorf("tenant not found")
	}
	m.tenants[tenant.ID] = &tenant
	return nil
}

func (m *MockTenantService) GetTenant(id string) (*domain.Tenant, error) {
	if tenant, exists := m.tenants[id]; exists {
		return tenant, nil
	}
	return nil, fmt.Errorf("tenant not found")
}

func (m *MockTenantService) DeleteTenant(id string) error {
	if _, exists := m.tenants[id]; !exists {
		return fmt.Errorf("tenant not found")
	}
	delete(m.tenants, id)
	return nil
}

func (m *MockTenantService) ListTenants(page, pageSize int) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	for _, t := range m.tenants {
		tenants = append(tenants, *t)
	}
	return tenants, nil
}

func (m *MockTenantService) GetTenantByDomain(domain string) (*domain.Tenant, error) {
	for _, t := range m.tenants {
		if t.CustomDomain == domain {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tenant not found")
}

func TestHandleWebhook(t *testing.T) {
	// Create mock tenant service
	mockTenantService := NewMockTenantService()

	// Create test tenant
	testTenant := &domain.Tenant{
		ID:                   "test-tenant-1",
		Name:                 "Test Tenant",
		Email:                "test@example.com",
		StripeCustomerID:     "cus_test_123",
		StripeSubscriptionID: "sub_test_123",
	}

	// Set up expectations
	mockTenantService.CreateTenant(*testTenant)

	// Create test handler
	handler := NewHandler(nil, mockTenantService)

	// Create test subscription event
	event := stripe.Event{
		Type: "customer.subscription.created",
		Data: &stripe.EventData{
			Raw: json.RawMessage(`{
				"id": "sub_test_123",
				"metadata": {
					"tenant_id": "test-tenant-1"
				},
				"items": {
					"data": [{
						"price": {
							"nickname": "basic"
						}
					}]
				}
			}`),
		},
	}

	// Create test request
	eventJSON, _ := json.Marshal(event)
	req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer(eventJSON))
	req.Header.Set("Stripe-Signature", "test_signature")

	// Create test response recorder
	w := httptest.NewRecorder()

	// Set webhook secret
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")

	// Handle webhook
	handler.HandleWebhook(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Get updated tenant
	updatedTenant, err := mockTenantService.GetTenant("test-tenant-1")
	assert.NoError(t, err)
	assert.Equal(t, "sub_test_123", updatedTenant.StripeSubscriptionID)
}

func TestHandleSubscriptionCreated(t *testing.T) {
	// Create mock tenant service
	mockTenantService := NewMockTenantService()

	// Create test tenant
	testTenant := &domain.Tenant{
		ID:                   "test-tenant-1",
		Name:                 "Test Tenant",
		Email:                "test@example.com",
		StripeCustomerID:     "cus_test_123",
		StripeSubscriptionID: "sub_test_123",
	}

	// Set up expectations
	mockTenantService.CreateTenant(*testTenant)

	// Create test handler
	handler := NewHandler(nil, mockTenantService)

	// Create test subscription
	subscription := &stripe.Subscription{
		ID: "sub_test_123",
		Metadata: map[string]string{
			"tenant_id": "test-tenant-1",
		},
		Items: &stripe.SubscriptionItemList{
			Data: []*stripe.SubscriptionItem{
				{
					Price: &stripe.Price{
						Nickname: "basic",
					},
				},
			},
		},
	}

	// Handle subscription created
	handler.handleSubscriptionCreated(subscription)

	// Get updated tenant
	updatedTenant, err := mockTenantService.GetTenant("test-tenant-1")
	assert.NoError(t, err)
	assert.Equal(t, "sub_test_123", updatedTenant.StripeSubscriptionID)
}

func TestWebhookHandler_HandleWebhook(t *testing.T) {
	// TODO: Implement webhook handler tests
	t.Skip("Webhook handler tests not implemented yet")
}
