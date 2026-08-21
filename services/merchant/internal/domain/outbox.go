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
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type OutboxRepository interface {
	Insert(ctx context.Context, event *OutboxEvent) error
	FetchUnpublished(ctx context.Context, limit int32) ([]*OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
}
