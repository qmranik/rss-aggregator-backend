package models

// PaymentRequest represents the data required to create a payment intent.
type PaymentRequest struct {
	Amount   int64  `json:"amount"`   // Amount in the smallest currency unit (e.g., cents for USD)
	Currency string `json:"currency"` // Currency code in ISO 4217 format (e.g., "usd")
}

// RefundRequest represents the data required to create a refund.
type RefundRequest struct {
	ChargeID string `json:"charge_id"` // ID of the charge to be refunded
	Amount   int64  `json:"amount"`    // Amount to be refunded in the smallest currency unit
}

// CreatePaymentParams represents the parameters for storing payment information in the database.
type CreatePaymentParams struct {
	ChargeID string `json:"charge_id"` // Stripe charge ID
	Amount   int64  `json:"amount"`    // Amount in the smallest currency unit
}
