package domain

import (
	"context"

	"github.com/google/uuid"
)

// LedgerResult is what a transfer returns: the transaction, its two journal
// entries, and the resulting balances (paise) on both accounts.
type LedgerResult struct {
	Transaction *Transaction
	DebitEntry  *JournalEntry
	CreditEntry *JournalEntry
	FromBalance int64 // paise after the transfer
	ToBalance   int64 // paise after the transfer
}

// LedgerRepository performs the atomic double-entry write. It MUST be a single
// database transaction so a failure never leaves a half-written ledger.
type LedgerRepository interface {
	ExecuteTransfer(ctx context.Context, t *Transaction) (*LedgerResult, error)
	// ExecuteTransferWithRewardEvent behaves like ExecuteTransfer but also
	// queues evt in the outbox within the same database transaction, so the
	// event is published if and only if the money actually moved. t.ID and
	// evt.TransactionID must be set by the caller.
	ExecuteTransferWithRewardEvent(ctx context.Context, t *Transaction, evt *PaymentCompletedEvent) (*LedgerResult, error)
	// ExecuteRewardCredit credits amount paise from the reward pool to the
	// wallet, recorded as a completed transaction keyed by payoutID for
	// idempotency. Returns the payment transaction ID; replaying the same
	// payoutID returns the original transaction without double-crediting.
	ExecuteRewardCredit(ctx context.Context, poolAccountID, walletAccountID, payoutID uuid.UUID, amount int64) (uuid.UUID, error)
}