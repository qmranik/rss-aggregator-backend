package stripe

import (
	"github.com/stripe/stripe-go/v72/client"
)

// StripeClient wraps the Stripe API client
type StripeClient struct {
	Client *client.API
}

// NewStripeClient creates a new StripeClient instance
func NewStripeClient(secretKey string) *StripeClient {
	sc := &client.API{}
	sc.Init(secretKey, nil)
	return &StripeClient{
		Client: sc,
	}
}
