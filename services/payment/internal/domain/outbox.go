package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          uuid.UUID
	Topic       string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}

type OutboxRepository interface {
	FetchUnpublished(ctx context.Context, limit int32) ([]*OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
}
