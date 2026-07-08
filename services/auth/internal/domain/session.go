package domain

import (
	"context"
	"time"
)

type Store interface {
	UpgradeTokenVersion(ctx context.Context, userID string) error
	GetTokenVersion(ctx context.Context, userID string) (int64, error)
	ClaimRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) (bool, error)
}