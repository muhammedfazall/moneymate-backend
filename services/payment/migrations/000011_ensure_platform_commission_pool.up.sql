-- Guarantees the platform commission pool row exists exactly once.
-- Insurance for databases that already recorded "version 9" from the rewards
-- integration migration and therefore skip 000009_seed_reward_pool.

INSERT INTO payment.accounts (type, currency)
SELECT 'platform_commission_pool', 'INR'
WHERE NOT EXISTS (
    SELECT 1 FROM payment.accounts WHERE type = 'platform_commission_pool'
);
