package domain

import (
	"time"

	"github.com/google/uuid"
)

// RewardPaymentCompletedTopic is the Kafka topic the rewards service consumes.
const RewardPaymentCompletedTopic = "moneymate.payment.completed"

// PaymentCompletedEvent is emitted whenever a user transfer completes.
// The Recipient* fields identify who EARNS the cashback — always the payer,
// so RecipientID is the sender's user ID and RecipientAccountID is the
// sender's wallet account ID.
type PaymentCompletedEvent struct {
	EventID            uuid.UUID `json:"event_id"`
	EventType          string    `json:"event_type"`
	TransactionID      uuid.UUID `json:"transaction_id"`
	RecipientID        uuid.UUID `json:"recipient_id"`
	RecipientAccountID uuid.UUID `json:"recipient_account_id"`
	RecipientType      string    `json:"recipient_type"`
	AmountPaise        int64     `json:"amount_paise"`
	OccurredAt         time.Time `json:"occurred_at"`
}
