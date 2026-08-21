package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeWallet             AccountType = "wallet"
	AccountTypePod                AccountType = "pod"
	AccountTypeMerchantSettlement AccountType = "merchant_settlement"
	AccountTypeMerchantPayout     AccountType = "merchant_payout"
	AccountTypePlatformCommission AccountType = "platform_commission_pool"
	AccountTypeExternalSettlement AccountType = "external_settlement"
)

type Account struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	MerchantID *uuid.UUID
	Type       AccountType
	Currency   string
	Balance    int64 // paise
	Handle     *string
	Version    int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AccountRepository interface {
	Create(ctx context.Context, a *Account) (*Account, error)
	CreateWallet(ctx context.Context, a *Account) (*Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByHandle(ctx context.Context, handle string) (*Account, error)
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	GetTotalBalanceByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	AddBalance(ctx context.Context, id uuid.UUID, amount int64) error
	CreateExternalSettlementAccount(ctx context.Context) (*Account, error)
 	GetExternalSettlementAccount(ctx context.Context) (*Account, error)
	CreateRewardPoolAccount(ctx context.Context) (*Account, error)
	GetRewardPoolAccount(ctx context.Context) (*Account, error)
	GetSystemAccountByType(ctx context.Context, accountType AccountType) (*Account, error)
}