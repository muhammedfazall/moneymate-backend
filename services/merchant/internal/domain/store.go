package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store represents the core merchant entity.
type Store struct {
	ID                uuid.UUID
	OwnerID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	OwnerName    string
	ContactEmail string
	MobileNumber string
	LegalName         string
	DBAName           *string
	Type              string
	TaxID             *string
	RegisteredAddress string
	DisplayID         string
	VPA               string
	QRCodeBase64      string
	PasswordHash      string
	Status            string
	Plan              string
	LogoURL           string
}

// KYCDocument represents compliance data.
type KYCDocument struct {
	ID             uuid.UUID
	StoreID        uuid.UUID
	VerifiedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AadhaarNumber  string
	AadhaarDocURL  string
	ShopLicenseURL string
	IsVerified     bool
}

// MerchantRepository defines the strict data access contract.
type MerchantRepository interface {
	RegisterStore(ctx context.Context, store *Store) (*Store, error)
	SubmitKYC(ctx context.Context, kyc *KYCDocument) error
	GetStoreByID(ctx context.Context, id uuid.UUID) (*Store, error)
	GetStoreByEmail(ctx context.Context, email string) (*Store, error)
	UpdateStoreStatus(ctx context.Context, storeID uuid.UUID, status string) error
	GetPendingStores(ctx context.Context) ([]*Store, error)
	GetStoreProfileByStoreID(ctx context.Context, storeID uuid.UUID) (*Store, error)
	UpdateStoreProfileByStoreID(ctx context.Context, store *Store) (*Store, error)
}