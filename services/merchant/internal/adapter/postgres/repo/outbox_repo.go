package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/sqlc/generated"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/pgxtx"
)

type OutboxRepo struct {
	db *pgxpool.Pool
	q  *generated.Queries
}

func NewOutboxRepo(db *pgxpool.Pool) *OutboxRepo {
	return &OutboxRepo{db: db, q: generated.New(db)}
}

func (r *OutboxRepo) Insert(ctx context.Context, e *domain.OutboxEvent) error {
	q := r.q
	if tx, ok := pgxtx.FromContext(ctx); ok {
		q = r.q.WithTx(tx)
	}
	return q.InsertOutboxEvent(ctx, generated.InsertOutboxEventParams{
		ID:      e.ID,
		Topic:   e.Topic,
		Payload: e.Payload,
	})
}

func (r *OutboxRepo) FetchUnpublished(ctx context.Context, limit int32) ([]*domain.OutboxEvent, error) {
	rows, err := r.q.FetchUnpublishedOutboxEvents(ctx, limit)
	if err != nil {
		return nil, err
	}
	var res []*domain.OutboxEvent
	for _, row := range rows {
		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			publishedAt = &row.PublishedAt.Time
		}
		res = append(res, &domain.OutboxEvent{
			ID:          row.ID,
			Topic:       row.Topic,
			Payload:     row.Payload,
			PublishedAt: publishedAt,
			CreatedAt:   row.CreatedAt,
		})
	}
	return res, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkOutboxEventPublished(ctx, id)
}
