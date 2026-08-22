-- References the 'reward_pool' enum value added in 000010, so it must be a
-- separate transaction committed after that one.

CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_reward_pool_singleton
ON payment.accounts ((type))
WHERE type = 'reward_pool';