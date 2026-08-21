package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/domain"
)

type RewardUsecase interface {
	CreateRule(ctx context.Context, rule domain.RewardRule) (*domain.RewardRule, error)
	ListRules(ctx context.Context, limit, offset int32) ([]*domain.RewardRule, error)
	GetRule(ctx context.Context, id uuid.UUID) (*domain.RewardRule, error)
	UpdateRule(ctx context.Context, rule domain.RewardRule) (*domain.RewardRule, error)
	DeactivateRule(ctx context.Context, id uuid.UUID) (*domain.RewardRule, error)
	ProcessPaymentCompletedEvent(ctx context.Context, payload []byte) error
	ListMyPayouts(ctx context.Context, userID uuid.UUID, status *domain.PayoutStatus, limit, offset int32) ([]*domain.RewardPayout, error)
	ListPayoutsByTransaction(ctx context.Context, transactionID uuid.UUID) ([]*domain.RewardPayout, error)
	GetPayoutByID(ctx context.Context, id uuid.UUID) (*domain.RewardPayout, error)
	ReplayFailed(ctx context.Context) error
}

type PaymentCompletedEvent struct {
	EventID            string    `json:"event_id"`
	EventType          string    `json:"event_type"`
	TransactionID      uuid.UUID `json:"transaction_id"`
	RecipientID        uuid.UUID `json:"recipient_id"`
	RecipientAccountID uuid.UUID `json:"recipient_account_id"`
	RecipientType      string    `json:"recipient_type"`
	AmountPaise        int64     `json:"amount_paise"`
	OccurredAt         time.Time `json:"occurred_at"`
}

type rewardUsecase struct {
	repo          domain.RewardRepository
	paymentClient domain.PaymentClient
}

func NewRewardUsecase(repo domain.RewardRepository, paymentClient domain.PaymentClient) RewardUsecase {
	return &rewardUsecase{repo: repo, paymentClient: paymentClient}
}

func (uc *rewardUsecase) CreateRule(ctx context.Context, rule domain.RewardRule) (*domain.RewardRule, error) {
	if rule.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if rule.MinPercentageBPS < 0 {
		return nil, fmt.Errorf("min_percentage_bps must be >= 0")
	}
	if rule.MaxPercentageBPS < rule.MinPercentageBPS {
		return nil, fmt.Errorf("max_percentage_bps must be >= min_percentage_bps")
	}
	if rule.MaxPercentageBPS > 10000 {
		return nil, fmt.Errorf("max_percentage_bps must be <= 10000")
	}
	if rule.MaxPayoutAmountPaise <= 0 {
		return nil, fmt.Errorf("max_payout_amount_paise must be > 0")
	}
	return uc.repo.CreateRule(ctx, rule)
}

func (uc *rewardUsecase) ListRules(ctx context.Context, limit, offset int32) ([]*domain.RewardRule, error) {
	if limit <= 0 {
		limit = 50
	}
	return uc.repo.ListRules(ctx, limit, offset)
}

func (uc *rewardUsecase) GetRule(ctx context.Context, id uuid.UUID) (*domain.RewardRule, error) {
	return uc.repo.GetRule(ctx, id)
}

func (uc *rewardUsecase) UpdateRule(ctx context.Context, rule domain.RewardRule) (*domain.RewardRule, error) {
	return uc.repo.UpdateRule(ctx, rule)
}

func (uc *rewardUsecase) DeactivateRule(ctx context.Context, id uuid.UUID) (*domain.RewardRule, error) {
	return uc.repo.DeactivateRule(ctx, id)
}

func (uc *rewardUsecase) ProcessPaymentCompletedEvent(ctx context.Context, payload []byte) error {
	var event PaymentCompletedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	rule, err := uc.repo.GetActiveRule(ctx)
	if err != nil {
		log.Printf("no active reward rule, skipping event %s", event.EventID)
		return nil
	}

	calc := CalculateReward(event.AmountPaise, *rule, func(minBPS, maxBPS int32) int32 {
		return minBPS + rand.Int31n(maxBPS-minBPS+1)
	})
	if !calc.Eligible {
		log.Printf("reward not eligible for event %s: %s", event.EventID, calc.Reason)
		return nil
	}

	recipientType := domain.RecipientTypeUser
	if event.RecipientType == "merchant" {
		recipientType = domain.RecipientTypeMerchant
	}

	payout := domain.RewardPayout{
		TransactionID:       event.TransactionID,
		RecipientID:         event.RecipientID,
		RecipientAccountID:  event.RecipientAccountID,
		RecipientType:       recipientType,
		RuleID:              &rule.ID,
		OriginalAmountPaise: event.AmountPaise,
		RewardPercentageBPS: calc.PercentageBPS,
		RewardAmountPaise:   calc.RewardAmountPaise,
		Status:              domain.PayoutStatusPending,
		EventPayload:        payload,
	}

	inserted, err := uc.repo.InsertPayout(ctx, payout)
	if err != nil {
		if isUniqueViolation(err) {
			log.Printf("duplicate event %s, already processed", event.EventID)
			return nil
		}
		return fmt.Errorf("insert payout: %w", err)
	}

	paymentTxID, err := uc.paymentClient.ExecuteRewardPayout(ctx, inserted.ID, event.RecipientAccountID, calc.RewardAmountPaise)
	if err != nil {
		_, _ = uc.repo.MarkFailed(ctx, inserted.ID, err.Error())
		return fmt.Errorf("execute payout: %w", err)
	}

	_, err = uc.repo.MarkCompleted(ctx, inserted.ID, paymentTxID)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}

	log.Printf("reward payout completed: payout_id=%s payment_tx=%s", inserted.ID, paymentTxID)
	return nil
}

func (uc *rewardUsecase) ListMyPayouts(ctx context.Context, userID uuid.UUID, status *domain.PayoutStatus, limit, offset int32) ([]*domain.RewardPayout, error) {
	if limit <= 0 {
		limit = 50
	}
	return uc.repo.ListPayoutsByRecipient(ctx, userID, status, limit, offset)
}

func (uc *rewardUsecase) ListPayoutsByTransaction(ctx context.Context, transactionID uuid.UUID) ([]*domain.RewardPayout, error) {
	return uc.repo.ListPayoutsByTransaction(ctx, transactionID)
}

func (uc *rewardUsecase) GetPayoutByID(ctx context.Context, id uuid.UUID) (*domain.RewardPayout, error) {
	return uc.repo.GetPayoutByID(ctx, id)
}

func (uc *rewardUsecase) ReplayFailed(ctx context.Context) error {
	payouts, err := uc.repo.ListFailedPayouts(ctx, 10)
	if err != nil {
		return fmt.Errorf("list failed payouts: %w", err)
	}

	for _, p := range payouts {
		paymentTxID, err := uc.paymentClient.ExecuteRewardPayout(ctx, p.ID, p.RecipientAccountID, p.RewardAmountPaise)
		if err != nil {
			log.Printf("replay failed for payout %s: %v", p.ID, err)
			continue
		}
		_, _ = uc.repo.MarkCompleted(ctx, p.ID, paymentTxID)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
