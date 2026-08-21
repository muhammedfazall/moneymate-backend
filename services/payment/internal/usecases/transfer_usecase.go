package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type TransferInput struct {
	AuthenticatedUserID string
	ToHandle            string
	AmountPaise         int64
	IdempotencyKey      string
	Description         string
	CategoryID          *string
}

type TransferUsecase interface {
	Transfer(ctx context.Context, in TransferInput) (*domain.LedgerResult, error)
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
}

type transferUsecase struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	ledger       domain.LedgerRepository
	categories   domain.CategoryRepository
}

func NewTransferUsecase(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	ledger domain.LedgerRepository,
	categories domain.CategoryRepository, // NEW
) TransferUsecase {
	return &transferUsecase{accounts: accounts, transactions: transactions, ledger: ledger, categories: categories}
}
func (u *transferUsecase) Transfer(ctx context.Context, in TransferInput) (*domain.LedgerResult, error) {
	if in.AmountPaise <= 0 {
		return nil, apperrors.ErrInvalidInput
	}
	key := strings.TrimSpace(in.IdempotencyKey)
	if key == "" {
		return nil, apperrors.ErrInvalidInput
	}
	authUserID, err := uuid.Parse(in.AuthenticatedUserID)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	fromAcc, err := u.accounts.GetWalletByUserID(ctx, authUserID)
	if err != nil {
		return nil, fmt.Errorf("get sender wallet: %w", err)
	}
	fromID := fromAcc.ID

	handle := strings.TrimSpace(in.ToHandle)
	if handle == "" {
		return nil, apperrors.ErrInvalidInput
	}

	toAcc, err := u.accounts.GetByHandle(ctx, handle)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrInvalidInput
		}
		return nil, err
	}
	if toAcc.Type != domain.AccountTypeWallet && toAcc.Type != domain.AccountTypeMerchantSettlement {
    return nil, apperrors.ErrInvalidInput
}
	toID := toAcc.ID

	if fromID == toID {
		return nil, apperrors.ErrInvalidInput
	}

	existing, err := u.transactions.GetByIdempotencyKey(ctx, key, fromID)
	if err == nil && existing != nil {
		return u.replay(ctx, existing)
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	var categoryID *uuid.UUID
	if in.CategoryID != nil && *in.CategoryID != "" {
		cid, err := uuid.Parse(*in.CategoryID)
		if err != nil {
			return nil, apperrors.ErrInvalidInput
		}
		cat, err := u.categories.GetByID(ctx, cid)
		if err != nil || cat.UserID != authUserID {
			return nil, apperrors.ErrInvalidInput
		}
		categoryID = &cid
	}

	txID := uuid.New()
	result, err := u.ledger.ExecuteTransferWithRewardEvent(ctx, &domain.Transaction{
		ID:             txID,
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         in.AmountPaise,
		Status:         domain.TxStatusPending,
		IdempotencyKey: key,
		Description:    in.Description,
		CategoryID:     categoryID,
	}, &domain.PaymentCompletedEvent{
		EventID:            uuid.New(),
		EventType:          domain.RewardPaymentCompletedTopic,
		TransactionID:      txID,
		RecipientID:        authUserID, // the payer earns the cashback
		RecipientAccountID: fromID,     // credited back to the payer's wallet
		RecipientType:      "user",
		AmountPaise:        in.AmountPaise,
		OccurredAt:         time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrIdempotencyKeyUsed) {
			winner, getErr := u.transactions.GetByIdempotencyKey(ctx, key, fromID)
			if getErr != nil {
				return nil, getErr
			}
			return u.replay(ctx, winner)
		}
		return nil, err
	}
	return result, nil
}

func (u *transferUsecase) replay(ctx context.Context, t *domain.Transaction) (*domain.LedgerResult, error) {
	entries, err := u.transactions.GetEntriesByTransactionID(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	result := &domain.LedgerResult{Transaction: t}
	for _, e := range entries {
		switch e.Direction {
		case "debit":
			result.DebitEntry = e
		case "credit":
			result.CreditEntry = e
		}
	}
	if from, err := u.accounts.GetByID(ctx, t.FromAccountID); err == nil {
		result.FromBalance = from.Balance
	}
	if to, err := u.accounts.GetByID(ctx, t.ToAccountID); err == nil {
		result.ToBalance = to.Balance
	}
	return result, nil
}

func (u *transferUsecase) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	txID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	return u.transactions.GetByID(ctx, txID)
}
