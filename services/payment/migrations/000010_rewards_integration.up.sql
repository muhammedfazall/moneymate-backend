-- Rewards service integration: transactional outbox + reward pool account.

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

CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_reward_pool_singleton
ON payment.accounts ((type))
WHERE type = 'reward_pool';
