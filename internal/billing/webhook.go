package billing

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	service       *StripeService
	tenantService domain.TenantService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(service *StripeService, tenantService domain.TenantService) *WebhookHandler {
	return &WebhookHandler{
		service:       service,
		tenantService: tenantService,
	}
}

// HandleWebhook processes incoming Stripe webhook events
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		http.Error(w, "Webhook secret is not configured", http.StatusInternalServerError)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		http.Error(w, "Error verifying webhook signature", http.StatusBadRequest)
		return
	}

	// Handle the event
	switch event.Type {
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing subscription", http.StatusBadRequest)
			return
		}
		// Handle subscription created
		h.handleSubscriptionCreated(&subscription)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing subscription", http.StatusBadRequest)
			return
		}
		// Handle subscription updated
		h.handleSubscriptionUpdated(&subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing subscription", http.StatusBadRequest)
			return
		}
		// Handle subscription deleted
		h.handleSubscriptionDeleted(&subscription)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			http.Error(w, "Error parsing invoice", http.StatusBadRequest)
			return
		}
		// Handle successful payment
		h.handlePaymentSucceeded(&invoice)

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			http.Error(w, "Error parsing invoice", http.StatusBadRequest)
			return
		}
		// Handle failed payment
		h.handlePaymentFailed(&invoice)

	default:
		// Unhandled event type
		http.Error(w, "Unhandled event type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleSubscriptionCreated handles the customer.subscription.created event
func (h *WebhookHandler) handleSubscriptionCreated(subscription *stripe.Subscription) {
	// Get tenant ID from metadata
	tenantID := subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		return
	}

	// Update tenant with Stripe subscription ID
	tenant.StripeSubscriptionID = subscription.ID
	tenant.Status = domain.StatusActive
	tenant.Plan = subscription.Items.Data[0].Price.Nickname // Assuming price nickname matches plan name

	// Update tenant in database
	h.tenantService.UpdateTenant(*tenant)
}

// handleSubscriptionUpdated handles the customer.subscription.updated event
func (h *WebhookHandler) handleSubscriptionUpdated(subscription *stripe.Subscription) {
	// Get tenant ID from metadata
	tenantID := subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		return
	}

	// Update tenant plan
	tenant.Plan = subscription.Items.Data[0].Price.Nickname // Assuming price nickname matches plan name

	// Update tenant in database
	h.tenantService.UpdateTenant(*tenant)
}

// handleSubscriptionDeleted handles the customer.subscription.deleted event
func (h *WebhookHandler) handleSubscriptionDeleted(subscription *stripe.Subscription) {
	// Get tenant ID from metadata
	tenantID := subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		return
	}

	// Update tenant status
	tenant.Status = domain.StatusInactive
	tenant.StripeSubscriptionID = ""

	// Update tenant in database
	h.tenantService.UpdateTenant(*tenant)
}

// handlePaymentSucceeded handles the invoice.payment_succeeded event
func (h *WebhookHandler) handlePaymentSucceeded(invoice *stripe.Invoice) {
	// Get tenant ID from subscription metadata
	subscription, err := subscription.Get(invoice.Subscription.ID, nil)
	if err != nil {
		return
	}

	tenantID := subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		return
	}

	// Update tenant status if needed
	if tenant.Status == domain.StatusSuspended {
		tenant.Status = domain.StatusActive
		h.tenantService.UpdateTenant(*tenant)
	}
}

// handlePaymentFailed handles the invoice.payment_failed event
func (h *WebhookHandler) handlePaymentFailed(invoice *stripe.Invoice) {
	// Get tenant ID from subscription metadata
	subscription, err := subscription.Get(invoice.Subscription.ID, nil)
	if err != nil {
		return
	}

	tenantID := subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		return
	}

	// Update tenant status
	tenant.Status = domain.StatusSuspended
	h.tenantService.UpdateTenant(*tenant)

	// TODO: Send notification to tenant about payment failure
}
