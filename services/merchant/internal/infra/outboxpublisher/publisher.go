
package outboxpublisher

import (
	"context"
	"log"
	"time"

	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/domain"
)

// KafkaProducer is the minimal contract this publisher needs — satisfied by
// whatever Kafka client library you pick (segmentio/kafka-go, confluent-kafka-go, etc).
type KafkaProducer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}

type Publisher struct {
	outbox   domain.OutboxRepository
	producer KafkaProducer
	interval time.Duration
	batch    int32
}

func New(outbox domain.OutboxRepository, producer KafkaProducer) *Publisher {
	return &Publisher{outbox: outbox, producer: producer, interval: 2 * time.Second, batch: 20}
}

func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.publishBatch(ctx)
		}
	}
}

func (p *Publisher) publishBatch(ctx context.Context) {
	events, err := p.outbox.FetchUnpublished(ctx, p.batch)
	if err != nil {
		log.Printf("outbox: fetch unpublished failed: %v", err)
		return
	}
	for _, e := range events {
		if err := p.producer.Publish(ctx, e.Topic, []byte(e.ID.String()), e.Payload); err != nil {
			log.Printf("outbox: publish event %s failed: %v", e.ID, err)
			continue // leave unpublished, retry next tick
		}
		if err := p.outbox.MarkPublished(ctx, e.ID); err != nil {
			log.Printf("outbox: mark published failed for %s: %v", e.ID, err)
		}
	}
}