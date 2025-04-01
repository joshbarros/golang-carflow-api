package billing

import (
	"os"
	"testing"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
)

func TestNewStripeService(t *testing.T) {
	// Test with missing secret key
	os.Unsetenv("STRIPE_SECRET_KEY")
	_, err := NewStripeService()
	assert.Error(t, err)
	assert.Equal(t, ErrStripeConfigMissing, err)

	// Test with valid secret key
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		t.Skip("STRIPE_SECRET_KEY not set")
	}
	os.Setenv("STRIPE_SECRET_KEY", secretKey)
	service, err := NewStripeService()
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestCreateCustomer(t *testing.T) {
	// Skip this test in CI/CD since it requires a real Stripe API key
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	// Use the secret key from environment
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		t.Skip("STRIPE_SECRET_KEY not set")
	}

	os.Setenv("STRIPE_SECRET_KEY", secretKey)
	service, err := NewStripeService()
	assert.NoError(t, err)

	tenant := &domain.Tenant{
		ID:    "test-tenant-1",
		Name:  "Test Tenant",
		Email: "test@example.com",
	}

	customer, err := service.CreateCustomer(tenant)
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, tenant.Email, customer.Email)
	assert.Equal(t, tenant.Name, customer.Name)
	assert.Equal(t, tenant.ID, customer.Metadata["tenant_id"])

	// Note: We don't clean up the customer in the test as it's better to keep test data
	// for debugging purposes. In a real environment, you would want to clean up after tests.
}

func TestCreateSubscription(t *testing.T) {
	// Skip this test in CI/CD since it requires a real Stripe API key
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	// Use the secret key from environment
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		t.Skip("STRIPE_SECRET_KEY not set")
	}

	os.Setenv("STRIPE_SECRET_KEY", secretKey)
	os.Setenv("STRIPE_PRICE_ID_BASIC", "price_basic_monthly")
	service, err := NewStripeService()
	assert.NoError(t, err)

	// Create a test customer first
	tenant := &domain.Tenant{
		ID:    "test-tenant-1",
		Name:  "Test Tenant",
		Email: "test@example.com",
	}

	customer, err := service.CreateCustomer(tenant)
	assert.NoError(t, err)
	tenant.StripeCustomerID = customer.ID

	// Create a test price
	params := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		Product:    stripe.String("prod_test"),
		UnitAmount: stripe.Int64(1000),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
	}
	price, err := price.New(params)
	assert.NoError(t, err)

	// Create subscription
	subscription, err := service.CreateSubscription(tenant, price.ID)
	assert.NoError(t, err)
	assert.NotNil(t, subscription)
	assert.Equal(t, tenant.ID, subscription.Metadata["tenant_id"])

	// Note: We don't clean up the subscription or customer in the test as it's better to keep test data
	// for debugging purposes. In a real environment, you would want to clean up after tests.
}

func TestGetPriceID(t *testing.T) {
	os.Setenv("STRIPE_PRICE_ID_BASIC", "price_basic_monthly")
	os.Setenv("STRIPE_PRICE_ID_PRO", "price_pro_monthly")
	os.Setenv("STRIPE_PRICE_ID_ENTERPRISE", "price_enterprise_monthly")
	service, err := NewStripeService()
	assert.NoError(t, err)

	tests := []struct {
		name    string
		plan    string
		want    string
		wantErr bool
	}{
		{
			name:    "basic plan",
			plan:    domain.PlanBasic,
			want:    "price_basic_monthly",
			wantErr: false,
		},
		{
			name:    "pro plan",
			plan:    domain.PlanPro,
			want:    "price_pro_monthly",
			wantErr: false,
		},
		{
			name:    "enterprise plan",
			plan:    domain.PlanEnterprise,
			want:    "price_enterprise_monthly",
			wantErr: false,
		},
		{
			name:    "invalid plan",
			plan:    "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetPriceID(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidPriceID, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
