package domain

import (
	"context"

	"github.com/google/uuid"
)

type PaymentClient interface {
	// ExecuteRewardPayout credits amountPaise to recipientAccountID.
	// payoutID is passed through for end-to-end idempotency: replaying the
	// same payout must never double-credit.
	ExecuteRewardPayout(ctx context.Context, payoutID, recipientAccountID uuid.UUID, amountPaise int64) (txID uuid.UUID, err error)
}
