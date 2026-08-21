package usecases_test

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
// 	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/usecases"
// )

// // Mock representations for dependencies

// type mockAccountRepo struct {
// 	accounts map[string]*domain.Account
// }

// func (m *mockAccountRepo) GetByHandle(ctx context.Context, handle string) (*domain.Account, error) {
// 	acc, ok := m.accounts[handle]
// 	if !ok {
// 		return nil, apperrors.ErrAccountNotFound
// 	}
// 	return acc, nil
// }
// func (m *mockAccountRepo) Create(ctx context.Context, account *domain.Account) error { return nil }
// func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) { return nil, nil }
// func (m *mockAccountRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Account, error) { return nil, nil }
// func (m *mockAccountRepo) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*domain.Account, error) { return nil, nil }
// func (m *mockAccountRepo) UpdateBalance(ctx context.Context, id uuid.UUID, amount int64) error { return nil }
// func (m *mockAccountRepo) WithTx(tx interface{}) domain.AccountRepository { return m }

// type mockTxManager struct{}
// func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
// 	return fn(ctx)
// }

// type mockTransferRepo struct{}
// func (m *mockTransferRepo) Create(ctx context.Context, t *domain.Transaction) error { return nil }
// func (m *mockTransferRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) { return nil, nil }
// func (m *mockTransferRepo) GetByIdempotencyKey(ctx context.Context, key string, fromID uuid.UUID) (*domain.Transaction, error) { return nil, errors.New("not found") }
// func (m *mockTransferRepo) GetEntriesByTransactionID(ctx context.Context, txID uuid.UUID) ([]*domain.LedgerEntry, error) { return nil, nil }
// func (m *mockTransferRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
// func (m *mockTransferRepo) WithTx(tx interface{}) domain.TransactionRepository { return m }

// type mockLedgerRepo struct{}
// func (m *mockLedgerRepo) ExecuteTransfer(ctx context.Context, tx *domain.Transaction) (*domain.LedgerResult, error) { return nil, nil }

// type mockCategoryRepo struct{}
// func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) { return nil, nil }

// type mockAuthClient struct {
// 	names  map[string]string
// 	photos map[string]string
// 	delay  time.Duration
// }

// func (m *mockAuthClient) GetUserProfile(ctx context.Context, userID string) (string, string, error) {
// 	if m.delay > 0 {
// 		select {
// 		case <-time.After(m.delay):
// 		case <-ctx.Done():
// 			return "", "", ctx.Err()
// 		}
// 	}
// 	name, ok := m.names[userID]
// 	if !ok {
// 		return "", "", errors.New("user not found")
// 	}
// 	return name, m.photos[userID], nil
// }

// type mockMerchantClient struct {
// 	names map[string]string
// 	logos map[string]string
// 	delay time.Duration
// }

// func (m *mockMerchantClient) GetStoreProfile(ctx context.Context, storeID string) (string, string, error) {
// 	if m.delay > 0 {
// 		select {
// 		case <-time.After(m.delay):
// 		case <-ctx.Done():
// 			return "", "", ctx.Err()
// 		}
// 	}
// 	name, ok := m.names[storeID]
// 	if !ok {
// 		return "", "", errors.New("store not found")
// 	}
// 	return name, m.logos[storeID], nil
// }

// func setupUsecase(accs map[string]*domain.Account, auths map[string]string, merchants map[string]string) (usecases.TransferUsecase, *mockAuthClient, *mockMerchantClient) {
// 	repo := &mockAccountRepo{accounts: accs}
// 	trepo := &mockTransferRepo{}
// 	ledger := &mockLedgerRepo{}
// 	categories := &mockCategoryRepo{}
	
// 	authClient := &mockAuthClient{names: auths, photos: make(map[string]string)}
// 	merchantClient := &mockMerchantClient{names: merchants, logos: make(map[string]string)}

// 	uc := usecases.NewTransferUsecase(repo, trepo, ledger, categories, authClient, merchantClient)
// 	return uc, authClient, merchantClient
// }

// func TestResolveHandle_RealUser(t *testing.T) {
// 	uID := uuid.New()
// 	accs := map[string]*domain.Account{
// 		"user@moneymate": {Type: domain.AccountTypeWallet, UserID: &uID},
// 	}
// 	auths := map[string]string{
// 		uID.String(): "John Doe",
// 	}
// 	uc, _, _ := setupUsecase(accs, auths, nil)

// 	res, err := uc.ResolveHandle(context.Background(), "user@moneymate")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if res.AccountType != string(domain.AccountTypeWallet) {
// 		t.Errorf("expected Wallet, got %s", res.AccountType)
// 	}
// 	if res.DisplayName != "John Doe" {
// 		t.Errorf("expected John Doe, got %s", res.DisplayName)
// 	}
// }

// func TestResolveHandle_RealMerchant(t *testing.T) {
// 	mID := uuid.New()
// 	accs := map[string]*domain.Account{
// 		"store@moneymate": {Type: domain.AccountTypeMerchantSettlement, MerchantID: &mID},
// 	}
// 	merchants := map[string]string{
// 		mID.String(): "Store LLC",
// 	}
// 	uc, _, _ := setupUsecase(accs, nil, merchants)

// 	res, err := uc.ResolveHandle(context.Background(), "store@moneymate")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if res.AccountType != string(domain.AccountTypeMerchantSettlement) {
// 		t.Errorf("expected MerchantSettlement, got %s", res.AccountType)
// 	}
// 	if res.DisplayName != "Store LLC" {
// 		t.Errorf("expected Store LLC, got %s", res.DisplayName)
// 	}
// }

// func TestResolveHandle_GarbageHandle(t *testing.T) {
// 	uc, _, _ := setupUsecase(map[string]*domain.Account{}, nil, nil)
// 	_, err := uc.ResolveHandle(context.Background(), "garbage")
// 	if err == nil {
// 		t.Fatalf("expected ErrAccountNotFound, got nil")
// 	}
// }

// func TestResolveHandle_DegradedUser(t *testing.T) {
// 	uID := uuid.New()
// 	accs := map[string]*domain.Account{
// 		"user@moneymate": {Type: domain.AccountTypeWallet, UserID: &uID},
// 	}
// 	uc, authC, _ := setupUsecase(accs, nil, nil)
// 	authC.delay = 5 * time.Second

// 	// Create context with 10ms timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
// 	defer cancel()

// 	res, err := uc.ResolveHandle(ctx, "user@moneymate")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if res.DisplayName != "" {
// 		t.Errorf("expected fallback empty display name, got %s", res.DisplayName)
// 	}
// }

// func TestResolveHandle_DegradedMerchant(t *testing.T) {
// 	mID := uuid.New()
// 	accs := map[string]*domain.Account{
// 		"store@moneymate": {Type: domain.AccountTypeMerchantSettlement, MerchantID: &mID},
// 	}
// 	uc, _, merchC := setupUsecase(accs, nil, nil)
// 	merchC.delay = 5 * time.Second

// 	// Create context with 10ms timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
// 	defer cancel()

// 	res, err := uc.ResolveHandle(ctx, "store@moneymate")
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if res.DisplayName != "" {
// 		t.Errorf("expected fallback empty display name, got %s", res.DisplayName)
// 	}
// }
