-- name: CreatePayment :exec
INSERT INTO payments (stripe_charge_id, amount, currency, status, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetPaymentByStripeID :one
SELECT *
FROM payments
WHERE stripe_charge_id = $1;

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET status = $1
WHERE stripe_charge_id = $2;

-- name: CreateRefund :exec
INSERT INTO refunds (stripe_refund_id, amount, status, email)
VALUES ($1, $2, $3, $4);

-- name: GetRefundByStripeID :one
SELECT *
FROM refunds
WHERE stripe_refund_id = $1;
