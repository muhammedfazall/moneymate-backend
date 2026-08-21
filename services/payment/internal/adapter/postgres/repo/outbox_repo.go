package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/payment/sqlc/generated"
)

type OutboxRepo struct {
	pool *pgxpool.Pool
	q    *generated.Queries
}

func NewOutboxRepo(pool *pgxpool.Pool) *OutboxRepo {
	return &OutboxRepo{pool: pool, q: generated.New(pool)}
}

func (r *OutboxRepo) FetchUnpublished(ctx context.Context, limit int32) ([]*domain.OutboxEvent, error) {
	rows, err := r.q.FetchUnpublishedOutboxEvents(ctx, limit)
	if err != nil {
		return nil, mapDBErr(err)
	}
	events := make([]*domain.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		e := &domain.OutboxEvent{
			ID:        row.ID,
			Topic:     row.Topic,
			Payload:   row.Payload,
			CreatedAt: row.CreatedAt,
		}
		if row.PublishedAt.Valid {
			t := row.PublishedAt.Time
			e.PublishedAt = &t
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	return mapDBErr(r.q.MarkOutboxEventPublished(ctx, id))
}
