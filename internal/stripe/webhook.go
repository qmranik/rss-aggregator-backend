package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/webhook"
)

// PaymentHandler manages Stripe payment and refund operations.
type PaymentHandler struct {
	StripeClient *StripeClient
	DB           *database.Queries
}

// CreatePaymentIntent handles the creation of a new payment intent.
func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request, user database.User) {
	// Set Stripe API key from environment variable
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Decode the payment request payload
	var paymentRequest models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create a new PaymentIntent with Stripe
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(paymentRequest.Amount),
		Currency: stripe.String(paymentRequest.Currency),
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		log.Printf("Failed to create payment intent: %v", err)
		http.Error(w, "Failed to create payment intent", http.StatusInternalServerError)
		return
	}

	// Save payment details to the database
	err = h.DB.CreatePayment(r.Context(), database.CreatePaymentParams{
		StripeChargeID: pi.ID,
		Amount:         paymentRequest.Amount,
		Currency:       paymentRequest.Currency,
		Status:         "created",
		Email:          user.Email,
	})
	if err != nil {
		log.Printf("Failed to save payment details: %v", err)
		http.Error(w, "Failed to save payment details", http.StatusInternalServerError)
		return
	}

	// Respond with the client secret for the PaymentIntent
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clientSecret": pi.ClientSecret,
	})
}

// CreateRefund handles the creation of a new refund.
func (h *PaymentHandler) CreateRefund(w http.ResponseWriter, r *http.Request, user database.User) {
	// Set Stripe API key from environment variable
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Decode the refund request payload
	var refundRequest models.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&refundRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve payment details from the database
	_, err := h.DB.GetPaymentByStripeID(r.Context(), refundRequest.ChargeID)
	if err != nil {
		log.Printf("Failed to get payment details: %v", err)
		http.Error(w, "Failed to get payment details", http.StatusInternalServerError)
		return
	}

	// Create a new Refund with Stripe
	params := &stripe.RefundParams{
		Charge: stripe.String(refundRequest.ChargeID),
		Amount: stripe.Int64(refundRequest.Amount),
	}
	rfd, err := refund.New(params)
	if err != nil {
		log.Printf("Failed to create refund: %v", err)
		http.Error(w, "Failed to create refund", http.StatusInternalServerError)
		return
	}

	// Save refund details to the database
	err = h.DB.CreateRefund(r.Context(), database.CreateRefundParams{
		StripeRefundID: rfd.ID,
		Amount:         refundRequest.Amount,
		Status:         "created",
		Email:          user.Email,
	})
	if err != nil {
		log.Printf("Failed to save refund details: %v", err)
		http.Error(w, "Failed to save refund details", http.StatusInternalServerError)
		return
	}

	// Respond with the refund details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rfd)
}

// HandleWebhook handles incoming webhook events from Stripe.
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	// Read and parse the webhook payload
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		http.Error(w, "Failed to verify signature", http.StatusBadRequest)
		return
	}

	// Handle the event based on its type
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		h.handlePaymentSuccess(r.Context(), pi)

	case "charge.refunded":
		var refund stripe.Refund
		if err := json.Unmarshal(event.Data.Raw, &refund); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		h.handleRefundSuccess(r.Context(), refund)

	default:
		fmt.Printf("Unhandled event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// handlePaymentSuccess updates the payment status to "succeeded" in the database.
func (h *PaymentHandler) handlePaymentSuccess(ctx context.Context, pi stripe.PaymentIntent) {
	err := h.DB.UpdatePaymentStatus(ctx, database.UpdatePaymentStatusParams{
		Status:         "succeeded",
		StripeChargeID: pi.ID,
	})
	if err != nil {
		log.Printf("Failed to update payment status: %v", err)
	}
}

// handleRefundSuccess updates the refund status to "succeeded" in the database.
func (h *PaymentHandler) handleRefundSuccess(ctx context.Context, refund stripe.Refund) {
	err := h.DB.CreateRefund(ctx, database.CreateRefundParams{
		StripeRefundID: refund.ID,
		Amount:         refund.Amount,
		Status:         "succeeded",
	})
	if err != nil {
		log.Printf("Failed to save refund details: %v", err)
	}
}
