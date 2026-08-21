DROP TABLE IF EXISTS payment.outbox_events;
DROP INDEX IF EXISTS payment.idx_accounts_reward_pool_singleton;

-- Postgres cannot remove enum values; 'reward_pool' simply remains unused.
