package billing

import "errors"

var (
	// ErrStripeConfigMissing is returned when Stripe configuration is missing
	ErrStripeConfigMissing = errors.New("stripe configuration is missing")
	// ErrCustomerNotFound is returned when a customer is not found
	ErrCustomerNotFound = errors.New("customer not found")
	// ErrSubscriptionNotFound is returned when a subscription is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrInvalidPriceID is returned when an invalid price ID is provided
	ErrInvalidPriceID = errors.New("invalid price ID")
	// ErrPaymentFailed is returned when a payment fails
	ErrPaymentFailed = errors.New("payment failed")
)
