package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type WalletUsecase interface {
	CreateWallet(ctx context.Context, userID, handle string) (*domain.Account, error)
	GetWallet(ctx context.Context, userID string) (*domain.Account, error)
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	GetByHandle(ctx context.Context, handle string) (*domain.Account, error)
	GetTotalBalance(ctx context.Context, userID string) (int64, error)
	GetWalletWithTotal(ctx context.Context, userID string) (*WalletBalanceResponse, error)
	ListAccounts(ctx context.Context, userID uuid.UUID) ([]*domain.Account, error)
	// CreditReward credits a reward payout to a wallet account. Idempotent
	// on payoutID. Returns the payment transaction ID.
	CreditReward(ctx context.Context, accountID, payoutID uuid.UUID, amountPaise int64) (uuid.UUID, error)
}

type walletUsecase struct {
	accounts     domain.AccountRepository
	ledger       domain.LedgerRepository
	rewardPoolID uuid.UUID
}

func NewWalletUsecase(accounts domain.AccountRepository, ledger domain.LedgerRepository, rewardPoolID uuid.UUID) WalletUsecase {
	return &walletUsecase{accounts: accounts, ledger: ledger, rewardPoolID: rewardPoolID}
}
func (u *walletUsecase) ListAccounts(ctx context.Context, userID uuid.UUID) ([]*domain.Account, error) {
	return u.accounts.ListByUser(ctx, userID)
}

// CreateWallet is idempotent: a user only ever gets one wallet.
func (u *walletUsecase) CreateWallet(ctx context.Context, userID, handle string) (*domain.Account, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	if handle == "" {
		return nil, apperrors.ErrInvalidInput
	}

	existing, err := u.accounts.GetWalletByUserID(ctx, uid)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	created, err := u.accounts.CreateWallet(ctx, &domain.Account{
		UserID: &uid, Currency: "INR", Handle: &handle,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrAlreadyExists) {
			return u.accounts.GetWalletByUserID(ctx, uid)
		}
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return created, nil
}

func (u *walletUsecase) GetWallet(ctx context.Context, userID string) (*domain.Account, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	return u.accounts.GetWalletByUserID(ctx, uid)
}
type WalletBalanceResponse struct {
	Wallet       *domain.Account
	TotalBalance int64
}

func (u *walletUsecase) GetWalletWithTotal(ctx context.Context, userID string) (*WalletBalanceResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	wallet, err := u.accounts.GetWalletByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	total, err := u.accounts.GetTotalBalanceByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get total balance: %w", err)
	}
	return &WalletBalanceResponse{Wallet: wallet, TotalBalance: total}, nil
}

func (u *walletUsecase) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	accID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	return u.accounts.GetByID(ctx, accID)
}

func (u *walletUsecase) GetByHandle(ctx context.Context, handle string) (*domain.Account, error) {
	if handle == "" {
		return nil, apperrors.ErrInvalidInput
	}
	return u.accounts.GetByHandle(ctx, handle)
}

func (u *walletUsecase) GetTotalBalance(ctx context.Context, userID string) (int64, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, apperrors.ErrInvalidInput
	}
	return u.accounts.GetTotalBalanceByUser(ctx, uid)
}

func (u *walletUsecase) CreditReward(ctx context.Context, accountID, payoutID uuid.UUID, amountPaise int64) (uuid.UUID, error) {
	if accountID == uuid.Nil || payoutID == uuid.Nil {
		return uuid.Nil, apperrors.ErrInvalidInput
	}
	acc, err := u.accounts.GetByID(ctx, accountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get reward target account: %w", err)
	}
	if acc.Type != domain.AccountTypeWallet {
		return uuid.Nil, apperrors.ErrInvalidInput
	}
	return u.ledger.ExecuteRewardCredit(ctx, u.rewardPoolID, accountID, payoutID, amountPaise)
}