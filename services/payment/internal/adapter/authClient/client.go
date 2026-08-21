package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func New(baseURL, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		http: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

type verifyTxTokenRequest struct {
	Token         string `json:"token"`
	TransactionID string `json:"transaction_id"`
}

type verifyTxTokenResponse struct {
	Data struct {
		Valid         bool   `json:"valid"`
		UserID        string `json:"user_id"`
		TransactionID string `json:"transaction_id"`
	} `json:"data"`
}

func (c *Client) VerifyTransactionToken(ctx context.Context, token, transactionID string) (string, error) {
	body, err := json.Marshal(verifyTxTokenRequest{Token: token, TransactionID: transactionID})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/internal/auth/verify-transaction-token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call auth-svc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transaction token rejected: status %d", resp.StatusCode)
	}

	var out verifyTxTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if !out.Data.Valid {
		return "", fmt.Errorf("transaction token invalid")
	}
	return out.Data.UserID, nil
}

type getUserProfileResponse struct {
	Data struct {
		FullName          string `json:"full_name"`
		ProfilePictureURL string `json:"profile_picture_url"`
	} `json:"data"`
}

func (c *Client) GetUserProfile(ctx context.Context, userID string) (string, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet,
		c.baseURL+"/internal/auth/users/"+userID, nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("call auth-svc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("auth-svc returned status %d", resp.StatusCode)
	}

	var out getUserProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return out.Data.FullName, out.Data.ProfilePictureURL, nil
}