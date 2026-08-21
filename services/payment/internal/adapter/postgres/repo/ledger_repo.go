package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/payment/sqlc/generated"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type LedgerRepo struct {
	pool *pgxpool.Pool
}

func NewLedgerRepo(pool *pgxpool.Pool) *LedgerRepo {
	return &LedgerRepo{pool: pool}
}


func (r *LedgerRepo) ExecuteTransfer(ctx context.Context, t *domain.Transaction) (*domain.LedgerResult, error) {
	return r.executeTransfer(ctx, t, nil)
}

func (r *LedgerRepo) ExecuteTransferWithRewardEvent(ctx context.Context, t *domain.Transaction, evt *domain.PaymentCompletedEvent) (*domain.LedgerResult, error) {
	return r.executeTransfer(ctx, t, evt)
}

func (r *LedgerRepo) executeTransfer(ctx context.Context, t *domain.Transaction, evt *domain.PaymentCompletedEvent) (*domain.LedgerResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ledger tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if already committed

	q := generated.New(tx)

	// Lock both accounts in sorted-ID order so two concurrent transfers
	// A→B and B→A never deadlock.
	lock1, lock2 := t.FromAccountID, t.ToAccountID
	if lock2.String() < lock1.String() {
		lock1, lock2 = lock2, lock1
	}
	first, err := q.GetAccountByIDForUpdate(ctx, lock1)
	if err != nil {
		return nil, mapDBErr(err)
	}
	second, err := q.GetAccountByIDForUpdate(ctx, lock2)
	if err != nil {
		return nil, mapDBErr(err)
	}

	var from generated.GetAccountByIDForUpdateRow
	if first.ID == t.FromAccountID {
		from = first
	} else {
		from = second
	}

	if from.Balance < t.Amount {
		return nil, apperrors.ErrInsufficientFunds
	}

	now := time.Now().UTC()
	txID := t.ID
	if txID == uuid.Nil {
		txID = uuid.New()
		t.ID = txID
	}

inserted, err := q.InsertTransaction(ctx, generated.InsertTransactionParams{
	ID:             txID,
	FromAccountID:  t.FromAccountID,
	ToAccountID:    t.ToAccountID,
	Amount:         t.Amount,
	Column5:        generated.PaymentTxStatus(domain.TxStatusCompleted),
	IdempotencyKey: t.IdempotencyKey,
	Description:    &t.Description,
	CategoryID:     categoryIDToPgtype(t.CategoryID), // NEW — needs a nullable UUID helper
	CompletedAt:    pgtype.Timestamptz{Time: now, Valid: true},
})
	if err != nil {
		return nil, mapDBErr(err) // unique-violation ⇒ ErrIdempotencyKeyUsed, handled by usecase
	}

	// Double entry: one DEBIT on the sender, one equal CREDIT on the receiver.
	if err := q.InsertJournalEntry(ctx, generated.InsertJournalEntryParams{
		ID:            uuid.New(),
		TransactionID: txID,
		AccountID:     t.FromAccountID,
		Amount:        t.Amount,
		Column5:       generated.PaymentTxDirectionDebit,
	}); err != nil {
		return nil, mapDBErr(err)
	}
	if err := q.InsertJournalEntry(ctx, generated.InsertJournalEntryParams{
		ID:            uuid.New(),
		TransactionID: txID,
		AccountID:     t.ToAccountID,
		Amount:        t.Amount,
		Column5:       generated.PaymentTxDirectionCredit,
	}); err != nil {
		return nil, mapDBErr(err)
	}

	// Update balances (negative = debit, positive = credit).
	if err := q.AddBalance(ctx, generated.AddBalanceParams{ID: t.FromAccountID, Balance: -t.Amount}); err != nil {
		return nil, mapDBErr(err)
	}
	if err := q.AddBalance(ctx, generated.AddBalanceParams{ID: t.ToAccountID, Balance: t.Amount}); err != nil {
		return nil, mapDBErr(err)
	}

	// Queue the reward event in the SAME transaction — it is published
	// if and only if this commit succeeds.
	if evt != nil {
		payload, err := json.Marshal(evt)
		if err != nil {
			return nil, fmt.Errorf("marshal reward event: %w", err)
		}
		if err := q.InsertOutboxEvent(ctx, generated.InsertOutboxEventParams{
			ID:      evt.EventID,
			Topic:   domain.RewardPaymentCompletedTopic,
			Payload: payload,
		}); err != nil {
			return nil, mapDBErr(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ledger tx: %w", err)
	}

	// Read fresh balances for the response (outside the tx, so no locks held).
	result := &domain.LedgerResult{Transaction: rowToTransaction(generated.GetTransactionByIDRow(inserted))}
	if fromAcc, err := r.getAccount(ctx, t.FromAccountID); err == nil {
		result.FromBalance = fromAcc.Balance
	}
	if toAcc, err := r.getAccount(ctx, t.ToAccountID); err == nil {
		result.ToBalance = toAcc.Balance
	}
	return result, nil
}

func (r *LedgerRepo) getAccount(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	return NewAccountRepo(r.pool).GetByID(ctx, id)
}

// ExecuteRewardCredit moves reward money from the pool account to a wallet.
// Idempotent on payoutID: replays return the original transaction without
// touching balances again.
func (r *LedgerRepo) ExecuteRewardCredit(ctx context.Context, poolAccountID, walletAccountID, payoutID uuid.UUID, amount int64) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, apperrors.ErrInvalidInput
	}
	if poolAccountID == walletAccountID {
		return uuid.Nil, apperrors.ErrInvalidInput
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin reward credit tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if already committed

	q := generated.New(tx)

	existing, err := q.GetTransactionByIdempotencyKey(ctx, generated.GetTransactionByIdempotencyKeyParams{
		IdempotencyKey: payoutID.String(),
		FromAccountID:  poolAccountID,
	})
	if err == nil {
		return existing.ID, nil // already credited — idempotent replay
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, mapDBErr(err)
	}

	if _, err := q.GetAccountByIDForUpdate(ctx, walletAccountID); err != nil {
		return uuid.Nil, mapDBErr(err)
	}

	now := time.Now().UTC()
	txID := uuid.New()
	desc := "reward payout"
	inserted, err := q.InsertTransaction(ctx, generated.InsertTransactionParams{
		ID:             txID,
		FromAccountID:  poolAccountID,
		ToAccountID:    walletAccountID,
		Amount:         amount,
		Column5:        generated.PaymentTxStatus(domain.TxStatusCompleted),
		IdempotencyKey: payoutID.String(),
		Description:    &desc,
		CategoryID:     pgtype.UUID{},
		CompletedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return uuid.Nil, mapDBErr(err)
	}

	if err := q.InsertJournalEntry(ctx, generated.InsertJournalEntryParams{
		ID:            uuid.New(),
		TransactionID: txID,
		AccountID:     poolAccountID,
		Amount:        amount,
		Column5:       generated.PaymentTxDirectionDebit,
	}); err != nil {
		return uuid.Nil, mapDBErr(err)
	}
	if err := q.InsertJournalEntry(ctx, generated.InsertJournalEntryParams{
		ID:            uuid.New(),
		TransactionID: txID,
		AccountID:     walletAccountID,
		Amount:        amount,
		Column5:       generated.PaymentTxDirectionCredit,
	}); err != nil {
		return uuid.Nil, mapDBErr(err)
	}

	if err := q.AddBalance(ctx, generated.AddBalanceParams{ID: poolAccountID, Balance: -amount}); err != nil {
		return uuid.Nil, mapDBErr(err)
	}
	if err := q.AddBalance(ctx, generated.AddBalanceParams{ID: walletAccountID, Balance: amount}); err != nil {
		return uuid.Nil, mapDBErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit reward credit tx: %w", err)
	}

	return inserted.ID, nil
}

func categoryIDToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}