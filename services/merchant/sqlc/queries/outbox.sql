-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, topic, payload)
VALUES ($1, $2, $3);

-- name: FetchUnpublishedOutboxEvents :many
SELECT id, topic, payload, published_at, created_at
FROM outbox_events
WHERE published_at IS NULL
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events
SET published_at = NOW()
WHERE id = $1 AND published_at IS NULL;
