DROP INDEX IF EXISTS merchant.idx_wallet_transactions_txn_type;
DROP INDEX IF EXISTS merchant.idx_wallet_transactions_store_id_created;
DROP INDEX IF EXISTS merchant.idx_wallets_store_id;
DROP TABLE IF EXISTS merchant.wallet_transactions;
DROP TABLE IF EXISTS merchant.wallets;
DROP TYPE IF EXISTS merchant.wallet_txn_type;

DROP INDEX IF EXISTS merchant.idx_redemption_requests_store_id_status;
DROP INDEX IF EXISTS merchant.idx_reward_transactions_display_id;
DROP INDEX IF EXISTS merchant.idx_reward_transactions_store_id_created;
DROP INDEX IF EXISTS merchant.idx_reward_balances_store_id;
DROP TABLE IF EXISTS merchant.redemption_requests;
DROP TABLE IF EXISTS merchant.reward_transactions;
DROP TYPE IF EXISTS merchant.redemption_status;
DROP TYPE IF EXISTS merchant.reward_transaction_type;

DROP INDEX IF EXISTS merchant.idx_unique_milestone_payout;
DROP INDEX IF EXISTS merchant.idx_earnings_payout_requests_store_id;
DROP TABLE IF EXISTS merchant.earnings_payout_requests;
DROP TABLE IF EXISTS merchant.earnings_stats;

DROP INDEX IF EXISTS merchant.idx_qr_transactions_store_id_created;
DROP TABLE IF EXISTS merchant.qr_transactions;