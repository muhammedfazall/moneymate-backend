-- reward_pool joins platform_commission_pool and external_settlement as an
-- ownerless system account (seeded at payment startup, no user/merchant owner).

ALTER TABLE payment.accounts DROP CONSTRAINT chk_account_owner;
ALTER TABLE payment.accounts ADD CONSTRAINT chk_account_owner CHECK (
    user_id IS NOT NULL OR merchant_id IS NOT NULL
    OR type IN ('platform_commission_pool', 'external_settlement', 'reward_pool')
);