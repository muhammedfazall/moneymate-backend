DROP TABLE IF EXISTS payment.outbox_events;

-- Postgres cannot remove enum values; 'reward_pool' simply remains unused.