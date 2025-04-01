package billing

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/joshbarros/golang-carflow-api/internal/domain"
	"github.com/joshbarros/golang-carflow-api/internal/interfaces"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Handler handles billing-related HTTP requests
type Handler struct {
	tenantService interfaces.TenantService
	stripeService *StripeService
}

// NewHandler creates a new billing handler
func NewHandler(tenantService interfaces.TenantService, stripeService *StripeService) *Handler {
	return &Handler{
		tenantService: tenantService,
		stripeService: stripeService,
	}
}

// RegisterRoutes registers the billing routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Webhook endpoint
	mux.HandleFunc("POST /webhooks/stripe", h.HandleWebhook)

	// Customer endpoints
	mux.HandleFunc("POST /billing/customers", h.HandleCreateCustomer)
	mux.HandleFunc("GET /billing/customers/{id}", h.HandleGetCustomer)

	// Subscription endpoints
	mux.HandleFunc("POST /billing/subscriptions", h.HandleCreateSubscription)
	mux.HandleFunc("GET /billing/subscriptions/{id}", h.HandleGetSubscription)
	mux.HandleFunc("PUT /billing/subscriptions/{id}", h.HandleUpdateSubscription)
	mux.HandleFunc("DELETE /billing/subscriptions/{id}", h.HandleCancelSubscription)
}

// HandleWebhook processes incoming Stripe webhook events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
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
func (h *Handler) handleSubscriptionCreated(subscription *stripe.Subscription) {
	// TODO: Update tenant status and plan
}

// handleSubscriptionUpdated handles the customer.subscription.updated event
func (h *Handler) handleSubscriptionUpdated(subscription *stripe.Subscription) {
	// TODO: Update tenant plan and features
}

// handleSubscriptionDeleted handles the customer.subscription.deleted event
func (h *Handler) handleSubscriptionDeleted(subscription *stripe.Subscription) {
	// TODO: Update tenant status to inactive
}

// handlePaymentSucceeded handles the invoice.payment_succeeded event
func (h *Handler) handlePaymentSucceeded(invoice *stripe.Invoice) {
	// TODO: Update tenant billing status
}

// handlePaymentFailed handles the invoice.payment_failed event
func (h *Handler) handlePaymentFailed(invoice *stripe.Invoice) {
	// TODO: Update tenant billing status and send notification
}

// HandleCreateCustomer handles POST /billing/customers
func (h *Handler) HandleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	var tenant domain.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create Stripe customer
	customer, err := h.stripeService.CreateCustomer(&tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update tenant with Stripe customer ID
	tenant.StripeCustomerID = customer.ID
	if err := h.tenantService.UpdateTenant(tenant); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

// HandleGetCustomer handles GET /billing/customers/{id}
func (h *Handler) HandleGetCustomer(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get customer
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleCreateSubscription handles POST /billing/subscriptions
func (h *Handler) HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
		Plan     string `json:"plan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(req.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get price ID for plan
	priceID, err := h.stripeService.GetPriceID(req.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create subscription
	subscription, err := h.stripeService.CreateSubscription(tenant, priceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update tenant with subscription ID
	tenant.StripeSubscriptionID = subscription.ID
	if err := h.tenantService.UpdateTenant(*tenant); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// HandleGetSubscription handles GET /billing/subscriptions/{id}
func (h *Handler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing subscription ID", http.StatusBadRequest)
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get subscription
	subscription, err := h.stripeService.GetSubscription(tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// HandleUpdateSubscription handles PUT /billing/subscriptions/{id}
func (h *Handler) HandleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing subscription ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get price ID for plan
	priceID, err := h.stripeService.GetPriceID(req.Plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update subscription
	subscription, err := h.stripeService.UpdateSubscription(tenant, priceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// HandleCancelSubscription handles DELETE /billing/subscriptions/{id}
func (h *Handler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing subscription ID", http.StatusBadRequest)
		return
	}

	// Get tenant from database
	tenant, err := h.tenantService.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Cancel subscription
	subscription, err := h.stripeService.CancelSubscription(tenant)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}
