package merchantclient

import (
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

type getStoreProfileResponse struct {
	LegalName string `json:"legal_name"`
	LogoURL   string `json:"logo_url"`
}

func (c *Client) GetStoreProfile(ctx context.Context, storeID string) (string, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet,
		c.baseURL+"/internal/merchant/stores/"+storeID+"/profile", nil)
	if err != nil {
		return "", "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("call merchant-svc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("merchant-svc returned status %d", resp.StatusCode)
	}

	var out getStoreProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return out.LegalName, out.LogoURL, nil
}