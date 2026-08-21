package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type SystemTransferInput struct {
	FromAccountID  uuid.UUID // caller supplies directly — e.g. the reward pool's account ID
	ToAccountID    uuid.UUID // caller supplies directly — resolved already, no handle lookup
	AmountPaise    int64
	IdempotencyKey string
	Description    string
}

type SystemTransferUsecase interface {
	Transfer(ctx context.Context, in SystemTransferInput) (*domain.LedgerResult, error)
}

type systemTransferUsecase struct {
	ledger domain.LedgerRepository
}

func NewSystemTransferUsecase(ledger domain.LedgerRepository) SystemTransferUsecase {
	return &systemTransferUsecase{ledger: ledger}
}

func (u *systemTransferUsecase) Transfer(ctx context.Context, in SystemTransferInput) (*domain.LedgerResult, error) {
	if in.AmountPaise <= 0 {
		return nil, apperrors.ErrInvalidInput
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return nil, apperrors.ErrInvalidInput
	}
	if in.FromAccountID == in.ToAccountID {
		return nil, apperrors.ErrInvalidInput
	}

	return u.ledger.ExecuteTransfer(ctx, &domain.Transaction{
		FromAccountID:  in.FromAccountID,
		ToAccountID:    in.ToAccountID,
		Amount:         in.AmountPaise,
		Status:         domain.TxStatusPending,
		IdempotencyKey: in.IdempotencyKey,
		Description:    in.Description,
	})
}