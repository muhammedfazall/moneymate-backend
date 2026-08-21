package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestPaymentCompletedEventWireContract verifies the JSON field names match
// exactly what the rewards service unmarshals (services/rewards
// usecases.PaymentCompletedEvent): event_id, event_type, transaction_id,
// recipient_id, recipient_account_id, recipient_type, amount_paise,
// occurred_at.
func TestPaymentCompletedEventWireContract(t *testing.T) {
	event := PaymentCompletedEvent{
		EventID:            uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		EventType:          RewardPaymentCompletedTopic,
		TransactionID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		RecipientID:        uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		RecipientAccountID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		RecipientType:      "user",
		AmountPaise:        250000,
		OccurredAt:         time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	requiredKeys := []string{
		"event_id",
		"event_type",
		"transaction_id",
		"recipient_id",
		"recipient_account_id",
		"recipient_type",
		"amount_paise",
		"occurred_at",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required JSON field %q in payload: %s", key, payload)
		}
	}
	if len(raw) != len(requiredKeys) {
		t.Errorf("unexpected extra fields: got %d fields, want %d", len(raw), len(requiredKeys))
	}

	if got := raw["event_type"]; got != RewardPaymentCompletedTopic {
		t.Errorf("event_type = %v, want %v", got, RewardPaymentCompletedTopic)
	}
	if got := raw["amount_paise"]; got != float64(250000) {
		t.Errorf("amount_paise = %v, want 250000", got)
	}

	var decoded PaymentCompletedEvent
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if decoded != event {
		t.Errorf("round-trip mismatch:\n got %+v\nwant %+v", decoded, event)
	}
}
