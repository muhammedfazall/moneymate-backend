package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/sqlc/generated"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/pgxtx"
)

// StoreRepo implements domain.MerchantRepository using pgxpool for mechanical sympathy.
type StoreRepo struct {
	db *pgxpool.Pool
	q  *generated.Queries
}

// NewStoreRepo initializes the repository instance.
func NewStoreRepo(db *pgxpool.Pool) *StoreRepo {
	return &StoreRepo{
		db: db,
		q:  generated.New(db),
	}
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// RegisterStore commits step 1 and 2 of the merchant onboarding flow.
func (r *StoreRepo) RegisterStore(ctx context.Context, store *domain.Store) (*domain.Store, error) {
	arg := generated.CreateStoreParams{
		ID:                store.ID,
		OwnerID:           store.OwnerID,
		OwnerName:         store.OwnerName,
		ContactEmail:      store.ContactEmail,
		MobileNumber:      store.MobileNumber,
		LegalName:         store.LegalName,
		DbaName:           store.DBAName,
		BusinessType:      store.Type, // store.Type is mapped to BusinessType string
		TaxID:             store.TaxID,
		RegisteredAddress: store.RegisteredAddress,
		DisplayID:         store.DisplayID,
		Vpa:               &store.VPA,
		QrCodeBase64:      &store.QRCodeBase64,
		PasswordHash:      store.PasswordHash,
	}

	q := r.q
	if tx, ok := pgxtx.FromContext(ctx); ok {
		q = r.q.WithTx(tx)
	}

	row, err := q.CreateStore(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.RegisterStore insertion failed: %w", err)
	}

	store.ID = row.ID
	store.Status = string(row.Status)
	store.Plan = string(row.Plan)
	store.VPA = safeString(row.Vpa)
	store.QRCodeBase64 = safeString(row.QrCodeBase64)
	store.CreatedAt = row.CreatedAt

	return store, nil
}

// SubmitKYC commits step 3 compliance documents.
func (r *StoreRepo) SubmitKYC(ctx context.Context, kyc *domain.KYCDocument) error {
	arg := generated.SubmitKYCParams{
		StoreID:        kyc.StoreID,
		AadhaarNumber:  kyc.AadhaarNumber,
		AadhaarDocUrl:  kyc.AadhaarDocURL,
		ShopLicenseUrl: kyc.ShopLicenseURL,
	}

	q := r.q
	if tx, ok := pgxtx.FromContext(ctx); ok {
		q = r.q.WithTx(tx)
	}

	if err := q.SubmitKYC(ctx, arg); err != nil {
		return fmt.Errorf("StoreRepo.SubmitKYC failed: %w", err)
	}
	return nil
}

// GetStoreByID retrieves the store state for gateway routing.
func (r *StoreRepo) GetStoreByID(ctx context.Context, id uuid.UUID) (*domain.Store, error) {
	row, err := r.q.GetStoreByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.GetStoreByID query failed: %w", err)
	}

	return &domain.Store{
		ID:        row.ID,
		DisplayID: row.DisplayID,
		VPA:       safeString(row.Vpa),
		LegalName: row.LegalName,
		Status:    string(row.Status),
		Plan:      string(row.Plan),
	}, nil
}

// GetStoreByEmail retrieves a store by its contact email.
func (r *StoreRepo) GetStoreByEmail(ctx context.Context, email string) (*domain.Store, error) {
	row, err := r.q.GetStoreByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.GetStoreByEmail query failed: %w", err)
	}

	return &domain.Store{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		DisplayID:    row.DisplayID,
		VPA:          safeString(row.Vpa),
		LegalName:    row.LegalName,
		Status:       string(row.Status),
		Plan:         string(row.Plan),
		PasswordHash: row.PasswordHash,
	}, nil
}

// UpdateStoreStatus advances the state machine (e.g., pending_kyc -> active).
func (r *StoreRepo) UpdateStoreStatus(ctx context.Context, storeID uuid.UUID, status string) error {
	arg := generated.UpdateStoreStatusParams{
		ID:     storeID,
		Status: generated.MerchantStatus(status),
	}

	if err := r.q.UpdateStoreStatus(ctx, arg); err != nil {
		return fmt.Errorf("StoreRepo.UpdateStoreStatus failed: %w", err)
	}
	return nil
}

// GetPendingStores retrieves all merchants in the 'pending_kyc' status.
func (r *StoreRepo) GetPendingStores(ctx context.Context) ([]*domain.Store, error) {
	rows, err := r.q.GetPendingStores(ctx)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.GetPendingStores failed: %w", err)
	}

	var stores []*domain.Store
	for _, row := range rows {
		stores = append(stores, &domain.Store{
			ID:           row.ID,
			OwnerName:    row.OwnerName,
			ContactEmail: row.ContactEmail,
			MobileNumber: row.MobileNumber,
			LegalName:    row.LegalName,
			Type:         row.BusinessType,
			Status:       string(row.Status),
			CreatedAt:    row.CreatedAt,
		})
	}
	return stores, nil
}

// GetStoreProfileByStoreID retrieves the complete merchant profile by store ID.
func (r *StoreRepo) GetStoreProfileByStoreID(ctx context.Context, storeID uuid.UUID) (*domain.Store, error) {
	row, err := r.q.GetStoreProfileByStoreID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.GetStoreProfileByStoreID failed: %w", err)
	}

	dba := row.DbaName
	tax := row.TaxID
	return &domain.Store{
		ID:                row.ID,
		OwnerID:           row.OwnerID,
		OwnerName:         row.OwnerName,
		ContactEmail:      row.ContactEmail,
		MobileNumber:      row.MobileNumber,
		LegalName:         row.LegalName,
		DBAName:           &dba,
		Type:              row.BusinessType,
		TaxID:             &tax,
		RegisteredAddress: row.RegisteredAddress,
		DisplayID:         row.DisplayID,
		VPA:               safeString(row.Vpa),
		QRCodeBase64:      safeString(row.QrCodeBase64),
		Status:            row.Status,
		Plan:              row.Plan,
		LogoURL:           row.LogoUrl,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}



// UpdateStoreProfileByStoreID updates the merchant profile using store ID.
func (r *StoreRepo) UpdateStoreProfileByStoreID(ctx context.Context, store *domain.Store) (*domain.Store, error) {
	var dba, tax string
	if store.DBAName != nil {
		dba = *store.DBAName
	}
	if store.TaxID != nil {
		tax = *store.TaxID
	}

	arg := generated.UpdateStoreProfileByStoreIDParams{
		StoreID:           store.ID,
		LegalName:         store.LegalName,
		DbaName:           dba,
		RegisteredAddress: store.RegisteredAddress,
		BusinessType:      store.Type,
		TaxID:             tax,
		OwnerName:         store.OwnerName,
		ContactEmail:      store.ContactEmail,
		MobileNumber:      store.MobileNumber,
		LogoUrl:           store.LogoURL,
	}

	row, err := r.q.UpdateStoreProfileByStoreID(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.UpdateStoreProfileByStoreID failed: %w", err)
	}

	resDba := row.DbaName
	resTax := row.TaxID
	return &domain.Store{
		ID:                row.ID,
		OwnerID:           row.OwnerID,
		OwnerName:         row.OwnerName,
		ContactEmail:      row.ContactEmail,
		MobileNumber:      row.MobileNumber,
		LegalName:         row.LegalName,
		DBAName:           &resDba,
		Type:              row.BusinessType,
		TaxID:             &resTax,
		RegisteredAddress: row.RegisteredAddress,
		DisplayID:         row.DisplayID,
		VPA:               safeString(row.Vpa),
		QRCodeBase64:      safeString(row.QrCodeBase64),
		Status:            row.Status,
		Plan:              row.Plan,
		LogoURL:           row.LogoUrl,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}
