-- Rewards service integration: transactional outbox + reward pool account.
-- NOTE: the singleton index lives in 000011 — Postgres forbids using an enum
-- value in the same transaction that adds it (error 55P04).

ALTER TYPE payment.account_type ADD VALUE IF NOT EXISTS 'reward_pool';

CREATE TABLE IF NOT EXISTS payment.outbox_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic        TEXT NOT NULL,
    payload      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_payment_outbox_unpublished
    ON payment.outbox_events (created_at)
    WHERE published_at IS NULL;