package subscription

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Handler handles HTTP requests for subscriptions
type Handler struct {
	stripeService *StripeService
	repo          Repository
}

// NewHandler creates a new subscription handler
func NewHandler(stripeService *StripeService, repo Repository) *Handler {
	return &Handler{
		stripeService: stripeService,
		repo:          repo,
	}
}

// CreateCheckoutSession handles POST /api/v1/subscriptions/checkout
func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant info from JWT context
	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email, ok := r.Context().Value("email").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Plan       string `json:"plan"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate plan
	if req.Plan != "starter" && req.Plan != "professional" && req.Plan != "enterprise" {
		respondError(w, "Invalid plan", http.StatusBadRequest)
		return
	}

	// Default URLs if not provided
	if req.SuccessURL == "" {
		req.SuccessURL = "http://localhost:3000/billing/success"
	}
	if req.CancelURL == "" {
		req.CancelURL = "http://localhost:3000/billing"
	}

	// Create checkout session
	checkoutReq := &CreateCheckoutSessionRequest{
		TenantID:      tenantID,
		TenantName:    "Tenant", // TODO: Get from tenant service
		CustomerEmail: email,
		CustomerName:  email,
		Plan:          req.Plan,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
	}

	checkoutURL, err := h.stripeService.CreateCheckoutSession(checkoutReq)
	if err != nil{
		log.Printf("Failed to create checkout session: %v", err)
		respondError(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"checkout_url": checkoutURL,
	}

	respondJSON(w, response, http.StatusOK)
}

// GetSubscription handles GET /api/v1/subscriptions/current
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant ID from JWT context
	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get subscription for tenant
	subscription, err := h.repo.GetByTenantID(tenantID)
	if err != nil {
		log.Printf("Failed to get subscription: %v", err)
		respondError(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	if subscription == nil {
		respondError(w, "No subscription found", http.StatusNotFound)
		return
	}

	respondJSON(w, subscription, http.StatusOK)
}

// CancelSubscription handles POST /api/v1/subscriptions/cancel
func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant ID from JWT context
	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CancelImmediately bool `json:"cancel_immediately"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get subscription for tenant
	subscription, err := h.repo.GetByTenantID(tenantID)
	if err != nil {
		log.Printf("Failed to get subscription: %v", err)
		respondError(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	if subscription == nil {
		respondError(w, "No subscription found", http.StatusNotFound)
		return
	}

	// Cancel subscription in Stripe
	err = h.stripeService.CancelSubscription(subscription.StripeSubscriptionID, req.CancelImmediately)
	if err != nil {
		log.Printf("Failed to cancel subscription: %v", err)
		respondError(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	message := "Subscription will be canceled at the end of the billing period"
	if req.CancelImmediately {
		message = "Subscription canceled immediately"
	}

	response := map[string]string{
		"message": message,
	}

	respondJSON(w, response, http.StatusOK)
}

// StripeWebhook handles POST /api/v1/webhooks/stripe
func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Get Stripe signature header
	signature := r.Header.Get("Stripe-Signature")

	// Handle webhook event
	err = h.stripeService.HandleWebhookEvent(payload, signature)
	if err != nil {
		log.Printf("Failed to handle webhook: %v", err)
		respondError(w, "Webhook error", http.StatusBadRequest)
		return
	}

	// Return 200 OK to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received": true}`))
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	respondJSON(w, response, statusCode)
}
