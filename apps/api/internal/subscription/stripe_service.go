package subscription

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/sub"
	"github.com/stripe/stripe-go/v81/webhook"
)

// StripeService handles Stripe payment operations
type StripeService struct {
	repo               Repository
	emailService       EmailService
	webhookSecret      string
	priceIDStarter     string
	priceIDPro         string
	priceIDEnterprise  string
}

// EmailService interface for sending emails
type EmailService interface {
	SendPaymentSuccessEmail(to, firstName string, amount float64, planName string) error
}

// NewStripeService creates a new Stripe service
func NewStripeService(repo Repository, emailService EmailService) *StripeService {
	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeService{
		repo:               repo,
		emailService:       emailService,
		webhookSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
		priceIDStarter:     getEnv("STRIPE_PRICE_STARTER", "price_starter"),
		priceIDPro:         getEnv("STRIPE_PRICE_PROFESSIONAL", "price_professional"),
		priceIDEnterprise:  getEnv("STRIPE_PRICE_ENTERPRISE", "price_enterprise"),
	}
}

// CreateCheckoutSessionRequest contains checkout session parameters
type CreateCheckoutSessionRequest struct {
	TenantID       string
	TenantName     string
	CustomerEmail  string
	CustomerName   string
	Plan           string
	SuccessURL     string
	CancelURL      string
}

// CreateCheckoutSession creates a Stripe checkout session
func (s *StripeService) CreateCheckoutSession(req *CreateCheckoutSessionRequest) (string, error) {
	if stripe.Key == "" {
		return "", fmt.Errorf("Stripe API key not configured")
	}

	// Get price ID for plan
	priceID := s.getPriceIDForPlan(req.Plan)

	// Create or get Stripe customer
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(req.CustomerEmail),
		Name:  stripe.String(req.CustomerName),
		Metadata: map[string]string{
			"tenant_id":   req.TenantID,
			"tenant_name": req.TenantName,
		},
	}

	cust, err := customer.New(customerParams)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	log.Printf("✅ Stripe customer created: %s for %s", cust.ID, req.CustomerEmail)

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(cust.ID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Metadata: map[string]string{
			"tenant_id": req.TenantID,
			"plan":      req.Plan,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"tenant_id":   req.TenantID,
				"tenant_name": req.TenantName,
				"plan":        req.Plan,
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	log.Printf("✅ Checkout session created: %s for tenant %s", sess.ID, req.TenantID)

	return sess.URL, nil
}

// HandleWebhookEvent processes Stripe webhook events
func (s *StripeService) HandleWebhookEvent(payload []byte, signature string) error {
	if s.webhookSecret == "" {
		log.Printf("⚠️  Stripe webhook secret not configured, skipping verification")
		// In development, you can skip verification, but log the event
		return nil
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	log.Printf("📥 Stripe webhook received: %s", event.Type)

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutSessionCompleted(event)
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(event)
	default:
		log.Printf("ℹ️  Unhandled webhook event type: %s", event.Type)
	}

	return nil
}

// handleCheckoutSessionCompleted handles successful checkout
func (s *StripeService) handleCheckoutSessionCompleted(event stripe.Event) error {
	var sess stripe.CheckoutSession
	if err := event.Data.Object.UnmarshalJSON(&sess); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	tenantID := sess.Metadata["tenant_id"]
	plan := sess.Metadata["plan"]

	log.Printf("✅ Checkout completed for tenant %s, plan: %s", tenantID, plan)

	// Create subscription record in database
	subscription := &Subscription{
		ID:                   uuid.New().String(),
		TenantID:             tenantID,
		StripeCustomerID:     sess.Customer.ID,
		StripeSubscriptionID: sess.Subscription.ID,
		Plan:                 plan,
		Status:               StatusActive,
	}

	err := s.repo.Create(subscription)
	if err != nil {
		log.Printf("❌ Failed to create subscription in database: %v", err)
		return err
	}

	log.Printf("✅ Subscription created in database for tenant %s", tenantID)
	return nil
}

// handleSubscriptionCreated handles new subscription
func (s *StripeService) handleSubscriptionCreated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := event.Data.Object.UnmarshalJSON(&subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	tenantID := subscription.Metadata["tenant_id"]
	plan := subscription.Metadata["plan"]

	log.Printf("✅ Subscription created: %s for tenant %s", subscription.ID, tenantID)

	// Check if subscription already exists
	existing, err := s.repo.GetByStripeSubscriptionID(subscription.ID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing
		existing.Status = string(subscription.Status)
		existing.CurrentPeriodStart = timePtr(time.Unix(subscription.CurrentPeriodStart, 0))
		existing.CurrentPeriodEnd = timePtr(time.Unix(subscription.CurrentPeriodEnd, 0))
		return s.repo.Update(existing)
	}

	// Create new subscription record
	newSub := &Subscription{
		ID:                   uuid.New().String(),
		TenantID:             tenantID,
		StripeCustomerID:     subscription.Customer.ID,
		StripeSubscriptionID: subscription.ID,
		StripePriceID:        subscription.Items.Data[0].Price.ID,
		Plan:                 plan,
		Status:               string(subscription.Status),
		CurrentPeriodStart:   timePtr(time.Unix(subscription.CurrentPeriodStart, 0)),
		CurrentPeriodEnd:     timePtr(time.Unix(subscription.CurrentPeriodEnd, 0)),
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
	}

	return s.repo.Create(newSub)
}

// handleSubscriptionUpdated handles subscription updates
func (s *StripeService) handleSubscriptionUpdated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := event.Data.Object.UnmarshalJSON(&subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	log.Printf("🔄 Subscription updated: %s, status: %s", subscription.ID, subscription.Status)

	// Get existing subscription
	existing, err := s.repo.GetByStripeSubscriptionID(subscription.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		log.Printf("⚠️  Subscription not found in database: %s", subscription.ID)
		return nil
	}

	// Update subscription
	existing.Status = string(subscription.Status)
	existing.CurrentPeriodStart = timePtr(time.Unix(subscription.CurrentPeriodStart, 0))
	existing.CurrentPeriodEnd = timePtr(time.Unix(subscription.CurrentPeriodEnd, 0))
	existing.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd

	if subscription.CanceledAt > 0 {
		existing.CanceledAt = timePtr(time.Unix(subscription.CanceledAt, 0))
	}

	return s.repo.Update(existing)
}

// handleSubscriptionDeleted handles subscription cancellation
func (s *StripeService) handleSubscriptionDeleted(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := event.Data.Object.UnmarshalJSON(&subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	log.Printf("❌ Subscription deleted: %s", subscription.ID)

	// Get existing subscription
	existing, err := s.repo.GetByStripeSubscriptionID(subscription.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	// Update status to canceled
	existing.Status = StatusCanceled
	existing.CanceledAt = timePtr(time.Now())

	return s.repo.Update(existing)
}

// handleInvoicePaymentSucceeded handles successful payment
func (s *StripeService) handleInvoicePaymentSucceeded(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := event.Data.Object.UnmarshalJSON(&invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	log.Printf("💰 Payment succeeded: %s, amount: $%.2f", invoice.ID, float64(invoice.AmountPaid)/100.0)

	// Get subscription
	if invoice.Subscription == nil {
		return nil
	}

	subscription, err := s.repo.GetByStripeSubscriptionID(invoice.Subscription.ID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return nil
	}

	// Send payment success email if email service is configured
	if s.emailService != nil {
		// We would need to get user email from tenant
		// For now, just log it
		log.Printf("📧 Would send payment success email for subscription %s", subscription.ID)
	}

	return nil
}

// handleInvoicePaymentFailed handles failed payment
func (s *StripeService) handleInvoicePaymentFailed(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := event.Data.Object.UnmarshalJSON(&invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	log.Printf("❌ Payment failed: %s", invoice.ID)

	// Get subscription and update status
	if invoice.Subscription == nil {
		return nil
	}

	subscription, err := s.repo.GetByStripeSubscriptionID(invoice.Subscription.ID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return nil
	}

	// Update subscription status
	subscription.Status = StatusPastDue
	return s.repo.Update(subscription)
}

// CancelSubscription cancels a subscription
func (s *StripeService) CancelSubscription(subscriptionID string, cancelImmediately bool) error {
	if stripe.Key == "" {
		return fmt.Errorf("Stripe API key not configured")
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(!cancelImmediately),
	}

	if cancelImmediately {
		_, err := sub.Cancel(subscriptionID, params)
		return err
	}

	_, err := sub.Update(subscriptionID, params)
	return err
}

// getPriceIDForPlan returns the Stripe price ID for a given plan
func (s *StripeService) getPriceIDForPlan(plan string) string {
	switch plan {
	case "starter":
		return s.priceIDStarter
	case "professional":
		return s.priceIDPro
	case "enterprise":
		return s.priceIDEnterprise
	default:
		return s.priceIDStarter
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
