package billing

import (
	"os"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/subscription"
)

// StripeService handles all Stripe-related operations
type StripeService struct {
	secretKey string
}

// NewStripeService creates a new Stripe service instance
func NewStripeService() (*StripeService, error) {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if secretKey == "" {
		return nil, ErrStripeConfigMissing
	}

	// Initialize Stripe client
	stripe.Key = secretKey

	return &StripeService{
		secretKey: secretKey,
	}, nil
}

// CreateCustomer creates a new Stripe customer for a tenant
func (s *StripeService) CreateCustomer(tenant *domain.Tenant) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(tenant.Email),
		Name:  stripe.String(tenant.Name),
		Params: stripe.Params{
			Metadata: map[string]string{
				"tenant_id": tenant.ID,
			},
		},
	}

	customer, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// CreateSubscription creates a new subscription for a tenant
func (s *StripeService) CreateSubscription(tenant *domain.Tenant, priceID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(tenant.StripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Params: stripe.Params{
			Metadata: map[string]string{
				"tenant_id": tenant.ID,
			},
		},
	}

	subscription, err := subscription.New(params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a tenant's subscription
func (s *StripeService) CancelSubscription(tenant *domain.Tenant) (*stripe.Subscription, error) {
	if tenant.StripeSubscriptionID == "" {
		return nil, ErrSubscriptionNotFound
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	subscription, err := subscription.Update(tenant.StripeSubscriptionID, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetSubscription retrieves a tenant's subscription
func (s *StripeService) GetSubscription(tenant *domain.Tenant) (*stripe.Subscription, error) {
	if tenant.StripeSubscriptionID == "" {
		return nil, ErrSubscriptionNotFound
	}

	subscription, err := subscription.Get(tenant.StripeSubscriptionID, nil)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscription updates a tenant's subscription
func (s *StripeService) UpdateSubscription(tenant *domain.Tenant, priceID string) (*stripe.Subscription, error) {
	if tenant.StripeSubscriptionID == "" {
		return nil, ErrSubscriptionNotFound
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		ProrationBehavior: stripe.String("always_invoice"),
	}

	subscription, err := subscription.Update(tenant.StripeSubscriptionID, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetPriceID returns the Stripe price ID for a given plan
func (s *StripeService) GetPriceID(plan string) (string, error) {
	switch plan {
	case domain.PlanBasic:
		return os.Getenv("STRIPE_PRICE_ID_BASIC"), nil
	case domain.PlanPro:
		return os.Getenv("STRIPE_PRICE_ID_PRO"), nil
	case domain.PlanEnterprise:
		return os.Getenv("STRIPE_PRICE_ID_ENTERPRISE"), nil
	default:
		return "", ErrInvalidPriceID
	}
}
