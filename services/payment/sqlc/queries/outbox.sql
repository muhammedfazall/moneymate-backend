-- name: InsertOutboxEvent :exec
INSERT INTO payment.outbox_events (id, topic, payload)
VALUES ($1, $2, $3);

-- name: FetchUnpublishedOutboxEvents :many
SELECT id, topic, payload, created_at, published_at
FROM payment.outbox_events
WHERE published_at IS NULL
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE payment.outbox_events
SET published_at = NOW()
WHERE id = $1;
