package paymentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HTTPClient executes reward payouts by calling the payment service's
// internal credit endpoint. The internal secret authenticates the call.
type HTTPClient struct {
	baseURL string
	secret  string
	client  *http.Client
}

func NewHTTPClient(baseURL, internalSecret string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  internalSecret,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type creditRewardRequest struct {
	PayoutID    string `json:"payout_id"`
	AmountPaise int64  `json:"amount_paise"`
}

type creditRewardResponse struct {
	TransactionID string `json:"transaction_id"`
}

func (c *HTTPClient) ExecuteRewardPayout(ctx context.Context, payoutID, recipientAccountID uuid.UUID, amountPaise int64) (uuid.UUID, error) {
	body, err := json.Marshal(creditRewardRequest{
		PayoutID:    payoutID.String(),
		AmountPaise: amountPaise,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal credit request: %w", err)
	}

	url := fmt.Sprintf("%s/internal/payment/wallets/%s/credit-reward", c.baseURL, recipientAccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return uuid.Nil, fmt.Errorf("build credit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("call payment service: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return uuid.Nil, fmt.Errorf("payment service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var out creditRewardResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return uuid.Nil, fmt.Errorf("decode credit response: %w", err)
	}
	txID, err := uuid.Parse(out.TransactionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid transaction id in response: %w", err)
	}
	return txID, nil
}
