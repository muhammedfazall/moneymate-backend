package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/domain"
)

type AdminRepo struct {
	db           *pgxpool.Pool
	storeRepo    *StoreRepo
	campaignRepo domain.CampaignRepository
	kycRepo      domain.KYCRepository
}

func NewAdminRepo(db *pgxpool.Pool, storeRepo *StoreRepo, campaignRepo domain.CampaignRepository, kycRepo domain.KYCRepository) domain.AdminRepository {
	return &AdminRepo{
		db:           db,
		storeRepo:    storeRepo,
		campaignRepo: campaignRepo,
		kycRepo:      kycRepo,
	}
}

// Stores
func (r *AdminRepo) GetAllStores(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			id, owner_name, contact_email, mobile_number,
			legal_name, COALESCE(dba_name, '') AS dba_name, business_type, COALESCE(tax_id, '') AS tax_id, registered_address,
			display_id, COALESCE(vpa, '') AS vpa, status::text, plan::text, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
		FROM stores
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetAllStores query failed: %w", err)
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		var s domain.Store
		var dba, tax string
		if err := rows.Scan(
			&s.ID, &s.OwnerName, &s.ContactEmail, &s.MobileNumber,
			&s.LegalName, &dba, &s.Type, &tax, &s.RegisteredAddress,
			&s.DisplayID, &s.VPA, &s.Status, &s.Plan, &s.LogoURL, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetAllStores scan failed: %w", err)
		}
		if dba != "" {
			s.DBAName = &dba
		}
		if tax != "" {
			s.TaxID = &tax
		}
		stores = append(stores, &s)
	}
	return stores, nil
}

func (r *AdminRepo) GetStoreByID(ctx context.Context, storeID uuid.UUID) (*domain.Store, error) {
	return r.storeRepo.GetStoreProfileByStoreID(ctx, storeID)
}

func (r *AdminRepo) UpdateStoreStatus(ctx context.Context, storeID uuid.UUID, status string) error {
	// Update in merchant DB
	err := r.storeRepo.UpdateStoreStatus(ctx, storeID, status)
	if err != nil {
		return err
	}

	// Fetch store to get email
	store, err := r.storeRepo.GetStoreProfileByStoreID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to fetch store profile for auth sync: %w", err)
	}

	// Sync with auth DB
	authStatus := "active"
	switch status {
	case "blocked", "suspended":
		authStatus = "suspended"
	case "deleted":
		authStatus = "deleted"
	case "pending_kyc":
		authStatus = "pending"
	}

	query := `UPDATE auth.users SET status = $1::text::auth.user_status, updated_at = NOW() WHERE email = $2;`
	_, err = r.db.Exec(ctx, query, authStatus, store.ContactEmail)
	if err != nil {
		// Just log or return error
		return fmt.Errorf("failed to sync status to auth.users: %w", err)
	}

	return nil
}

func (r *AdminRepo) DeleteStore(ctx context.Context, storeID uuid.UUID) error {
	query := `DELETE FROM stores WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, storeID)
	if err != nil {
		return fmt.Errorf("DeleteStore failed: %w", err)
	}
	return nil
}

// Campaigns
func (r *AdminRepo) GetAllCampaigns(ctx context.Context, limit, offset int) ([]*domain.Campaign, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			id, store_id, name, offer_type, COALESCE(reward_value, 0), COALESCE(min_bill_amount, 0), target_audience, start_date, end_date, status, created_at, updated_at
		FROM campaigns
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetAllCampaigns query failed: %w", err)
	}
	defer rows.Close()

	var campaigns []*domain.Campaign
	for rows.Next() {
		var c domain.Campaign
		if err := rows.Scan(
			&c.ID, &c.StoreID, &c.Name, &c.OfferType, &c.RewardValue, &c.MinBillAmount,
			&c.TargetAudience, &c.StartDate, &c.EndDate, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetAllCampaigns scan failed: %w", err)
		}
		campaigns = append(campaigns, &c)
	}
	return campaigns, nil
}

func (r *AdminRepo) CreateCampaign(ctx context.Context, c *domain.Campaign) (*domain.Campaign, error) {
	if c.Status == "" {
		c.Status = "active"
	}
	return r.campaignRepo.CreateCampaign(ctx, c)
}

func (r *AdminRepo) GetCampaignsByStoreID(ctx context.Context, storeID uuid.UUID) ([]*domain.Campaign, error) {
	return r.campaignRepo.GetCampaignsByStoreID(ctx, storeID)
}

func (r *AdminRepo) UpdateCampaignStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	status := "paused"
	if isActive {
		status = "active"
	}
	query := `UPDATE campaigns SET status = $2, updated_at = NOW() WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("UpdateCampaignStatus failed: %w", err)
	}
	return nil
}

func (r *AdminRepo) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM campaigns WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteCampaign failed: %w", err)
	}
	return nil
}

// KYC Verification
func (r *AdminRepo) GetAllKYCDocuments(ctx context.Context, limit, offset int) ([]*domain.KYCStatusDetail, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			k.id, k.store_id, COALESCE(k.aadhaar_number, ''), COALESCE(k.aadhaar_doc_url, ''), COALESCE(k.shop_license_url, ''),
			k.is_verified, k.verified_at, k.created_at, k.updated_at, COALESCE(s.status::text, 'pending_kyc')
		FROM kyc_documents k
		LEFT JOIN stores s ON k.store_id = s.id
		ORDER BY k.created_at DESC
		LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetAllKYCDocuments query failed: %w", err)
	}
	defer rows.Close()

	var docs []*domain.KYCStatusDetail
	for rows.Next() {
		var k domain.KYCStatusDetail
		var verifiedAt *time.Time
		if err := rows.Scan(
			&k.ID, &k.StoreID, &k.AadhaarNumber, &k.AadhaarDocURL, &k.ShopLicenseURL,
			&k.IsVerified, &verifiedAt, &k.CreatedAt, &k.UpdatedAt, &k.StoreStatus,
		); err != nil {
			return nil, fmt.Errorf("GetAllKYCDocuments scan failed: %w", err)
		}
		k.VerifiedAt = verifiedAt
		docs = append(docs, &k)
	}
	return docs, nil
}

func (r *AdminRepo) GetKYCByStoreID(ctx context.Context, storeID uuid.UUID) (*domain.KYCStatusDetail, error) {
	return r.kycRepo.GetKYCStatusByStoreID(ctx, storeID)
}

func (r *AdminRepo) VerifyKYCDocument(ctx context.Context, storeID uuid.UUID, isVerified bool, status string) (*domain.KYCStatusDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin VerifyKYCDocument tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var verifiedAt *time.Time
	now := time.Now().UTC()
	if isVerified {
		verifiedAt = &now
	}

	updateKYC := `
		UPDATE kyc_documents
		SET is_verified = $2, verified_at = $3, updated_at = NOW()
		WHERE store_id = $1;`
	_, err = tx.Exec(ctx, updateKYC, storeID, isVerified, verifiedAt)
	if err != nil {
		return nil, fmt.Errorf("update kyc_documents failed: %w", err)
	}

	if status != "" {
		updateStore := `UPDATE stores SET status = $2::text::merchant_status, updated_at = NOW() WHERE id = $1;`
		_, err = tx.Exec(ctx, updateStore, storeID, status)
		if err != nil {
			return nil, fmt.Errorf("update store status failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit VerifyKYCDocument tx: %w", err)
	}

	return r.GetKYCByStoreID(ctx, storeID)
}

// Rewards
func (r *AdminRepo) GetAllRewardTransactions(ctx context.Context, limit, offset int) ([]*domain.RewardTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			id, store_id, COALESCE(campaign_name, ''), COALESCE(display_id, ''), COALESCE(status, ''), COALESCE(amount, 0), COALESCE(transaction_type, ''), created_at
		FROM reward_transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetAllRewardTransactions query failed: %w", err)
	}
	defer rows.Close()

	var txs []*domain.RewardTransaction
	for rows.Next() {
		var t domain.RewardTransaction
		if err := rows.Scan(&t.ID, &t.StoreID, &t.CampaignName, &t.DisplayID, &t.Status, &t.Amount, &t.TransactionType, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetAllRewardTransactions scan failed: %w", err)
		}
		txs = append(txs, &t)
	}
	return txs, nil
}

func (r *AdminRepo) GetPlatformRewardSummary(ctx context.Context) (float64, int64, error) {
	query := `SELECT COALESCE(SUM(available_balance), 0), COALESCE(SUM(total_scans), 0) FROM reward_balances;`
	var totalBal float64
	var totalScans int64
	err := r.db.QueryRow(ctx, query).Scan(&totalBal, &totalScans)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, fmt.Errorf("GetPlatformRewardSummary failed: %w", err)
	}
	return totalBal, totalScans, nil
}

// Subscriptions
func (r *AdminRepo) GetAllSubscriptions(ctx context.Context, limit, offset int) ([]*domain.MerchantSubscription, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT 
			id, store_id, plan_code, COALESCE(status, 'active'), COALESCE(billing_cycle, 'monthly'), current_period_start, current_period_end, auto_renew, created_at, updated_at
		FROM merchant_subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetAllSubscriptions query failed: %w", err)
	}
	defer rows.Close()

	var subs []*domain.MerchantSubscription
	for rows.Next() {
		var s domain.MerchantSubscription
		if err := rows.Scan(
			&s.ID, &s.StoreID, &s.PlanCode, &s.Status, &s.BillingCycle,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetAllSubscriptions scan failed: %w", err)
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

func (r *AdminRepo) UpdateStoreSubscriptionPlan(ctx context.Context, storeID uuid.UUID, planCode string) (*domain.MerchantSubscription, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin UpdateStoreSubscriptionPlan tx: %w", err)
	}
	defer tx.Rollback(ctx)

	updateSub := `
		INSERT INTO merchant_subscriptions (store_id, plan_code, status, billing_cycle, current_period_start, current_period_end, auto_renew, created_at, updated_at)
		VALUES ($1, $2, 'active', 'monthly', NOW(), NOW() + INTERVAL '30 days', true, NOW(), NOW())
		ON CONFLICT (store_id) DO UPDATE 
		SET plan_code = EXCLUDED.plan_code, updated_at = NOW()
		RETURNING id, store_id, plan_code, status, billing_cycle, current_period_start, current_period_end, auto_renew, created_at, updated_at;`
	var s domain.MerchantSubscription
	err = tx.QueryRow(ctx, updateSub, storeID, planCode).Scan(
		&s.ID, &s.StoreID, &s.PlanCode, &s.Status, &s.BillingCycle,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update merchant_subscriptions failed: %w", err)
	}

	updateStore := `UPDATE stores SET plan = $2::text::subscription_plan, updated_at = NOW() WHERE id = $1;`
	_, err = tx.Exec(ctx, updateStore, storeID, planCode)
	if err != nil {
		return nil, fmt.Errorf("update store plan enum failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit UpdateStoreSubscriptionPlan tx: %w", err)
	}

	return &s, nil
}

// Dashboard
func (r *AdminRepo) GetAdminDashboardStats(ctx context.Context) (*domain.AdminDashboardStats, error) {
	stats := &domain.AdminDashboardStats{}

	err := r.db.QueryRow(ctx, `SELECT COALESCE(SUM(total_earnings), 0), COALESCE(SUM(available_balance), 0) FROM wallets;`).Scan(&stats.TotalRevenue, &stats.SystemWallet)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to fetch total revenue and wallet: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(available_balance), 0) FROM reward_balances;`).Scan(&stats.RewardPool)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to fetch reward pool: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM wallet_transactions WHERE created_at >= CURRENT_DATE;`).Scan(&stats.DailyTransactions)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to fetch daily transactions: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT transaction_id, title, amount, created_at, 'Completed' AS status 
		FROM wallet_transactions 
		ORDER BY created_at DESC LIMIT 5;`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tx domain.AdminRecentTransaction
		var t time.Time
		var amt float64
		if err := rows.Scan(&tx.ID, &tx.User, &amt, &t, &tx.Status); err == nil {
			tx.Amount = fmt.Sprintf("₹%.2f", amt)
			tx.Date = t.Format("2006-01-02")
			stats.RecentTransactions = append(stats.RecentTransactions, tx)
		}
	}

	if stats.RecentTransactions == nil {
		stats.RecentTransactions = []domain.AdminRecentTransaction{}
	}

	return stats, nil
}
