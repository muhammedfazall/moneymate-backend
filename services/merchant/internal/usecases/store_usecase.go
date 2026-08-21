package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/domain"
	hashpass "github.com/moneymate-2026/moneymate-backend/shared/pkg/hash"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/qrcode"
)

// StoreUseCase orchestrates merchant workflows.
type StoreUseCase struct {
	repo   domain.MerchantRepository
	outbox domain.OutboxRepository
	tx     domain.TxManager
}

// NewStoreUseCase constructs a usecase with repository dependencies.
func NewStoreUseCase(repo domain.MerchantRepository, outbox domain.OutboxRepository, tx domain.TxManager) *StoreUseCase {
	return &StoreUseCase{repo: repo, outbox: outbox, tx: tx}
}

type RegisterStoreInput struct {
	OwnerName         string
	ContactEmail      string
	MobileNumber      string
	LegalName         string
	DBAName           *string
	Type              string
	TaxID             *string
	RegisteredAddress string
	AadhaarNumber     string
	AadhaarDocURL     string
	ShopLicenseURL    string
	Password          string
}

// ProcessRegistration applies validation and executes state persistence.
func (uc *StoreUseCase) ProcessRegistration(ctx context.Context, in RegisterStoreInput) (*domain.Store, error) {
	storeIDStr, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate store UUID: %w", err)
	}
	storeID := storeIDStr

	displayID, err := generateDisplayID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure display ID: %w", err)
	}

	vpa := generateVPA(in.ContactEmail)
	qrPayload := qrcode.BuildPaymentPayload("merchant", vpa)
	qrCodeBase64, err := qrcode.GenerateBase64(qrPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	hashedPwd, err := hashpass.HashPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	store := &domain.Store{
		ID:                storeID,
		OwnerID:           storeID,
		OwnerName:         strings.TrimSpace(in.OwnerName),
		ContactEmail:      strings.ToLower(strings.TrimSpace(in.ContactEmail)),
		MobileNumber:      strings.TrimSpace(in.MobileNumber),
		LegalName:         strings.TrimSpace(in.LegalName),
		DBAName:           in.DBAName,
		Type:              in.Type,
		TaxID:             in.TaxID,
		RegisteredAddress: strings.TrimSpace(in.RegisteredAddress),
		DisplayID:         displayID,
		VPA:               vpa,
		QRCodeBase64:      qrCodeBase64,
		PasswordHash:      hashedPwd,
	}

	var createdStore *domain.Store
	err = uc.tx.WithTx(ctx, func(ctx context.Context) error {
		var txErr error
		createdStore, txErr = uc.repo.RegisterStore(ctx, store)
		if txErr != nil {
			return fmt.Errorf("failed to register store: %w", txErr)
		}

		kyc := &domain.KYCDocument{
			StoreID:        createdStore.ID,
			AadhaarNumber:  in.AadhaarNumber,
			AadhaarDocURL:  in.AadhaarDocURL,
			ShopLicenseURL: in.ShopLicenseURL,
		}

		if txErr := uc.repo.SubmitKYC(ctx, kyc); txErr != nil {
			return fmt.Errorf("failed to submit KYC documents: %w", txErr)
		}

		// Insert outbox event
		outboxID, txErr := uuid.NewV7()
		if txErr != nil {
			return fmt.Errorf("failed to generate outbox ID: %w", txErr)
		}

		payload, txErr := json.Marshal(map[string]string{
			"merchant_id": createdStore.ID.String(),
			"handle":      vpa,
		})
		if txErr != nil {
			return fmt.Errorf("marshal outbox payload: %w", txErr)
		}

		if txErr := uc.outbox.Insert(ctx, &domain.OutboxEvent{
			ID:      outboxID,
			Topic:   "merchant.registered",
			Payload: payload,
		}); txErr != nil {
			return fmt.Errorf("insert outbox event: %w", txErr)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdStore, nil
}

// GetStore retrieves a store by ID.
func (uc *StoreUseCase) GetStore(ctx context.Context, id string) (*domain.Store, error) {
	storeUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid store UUID format: %w", err)
	}

	return uc.repo.GetStoreByID(ctx, storeUUID)
}

// AuthenticateStore checks if the email and password match a registered store.
func (uc *StoreUseCase) AuthenticateStore(ctx context.Context, email, password string) (*domain.Store, error) {
	store, err := uc.repo.GetStoreByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	match, err := hashpass.VerifyPassword(store.PasswordHash, password)
	if err != nil || !match {
		return nil, fmt.Errorf("invalid credentials")
	}

	return store, nil
}

// generateDisplayID yields a collision-resistant MM-XXXX-XX identifier.
func generateDisplayID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hexStr := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("MM-%s-%s", hexStr[:4], hexStr[4:]), nil
}

// generateVPA creates a unique VPA like emailprefix+random@moneymate
func generateVPA(email string) string {
	parts := strings.Split(email, "@")
	prefix := parts[0]
	if len(prefix) > 10 {
		prefix = prefix[:10]
	}

	b := make([]byte, 2)
	rand.Read(b)
	hexStr := strings.ToLower(hex.EncodeToString(b))

	return fmt.Sprintf("%s%s@moneymate", prefix, hexStr)
}

// GetPendingStores retrieves all merchants in the pending_kyc status.
func (uc *StoreUseCase) GetStoreProfile(ctx context.Context, storeID uuid.UUID) (*domain.Store, error) {
	return uc.repo.GetStoreByID(ctx, storeID)
}

func (uc *StoreUseCase) GetPendingStores(ctx context.Context) ([]*domain.Store, error) {
	return uc.repo.GetPendingStores(ctx)
}

type UpdateProfileInput struct {
	StoreID      string
	BusinessName string
	DBAName      string
	Address      string
	BusinessType string
	TaxID        string
	OwnerName    string
	Email        string
	Mobile       string
	ProfileImage string
}

// GetProfile retrieves the store profile by store UUID.
func (uc *StoreUseCase) GetProfile(ctx context.Context, id string) (*domain.Store, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format for profile lookup: %w", err)
	}

	store, err := uc.repo.GetStoreProfileByStoreID(ctx, parsedUUID)
	if err == nil && store != nil && store.ID != uuid.Nil {
		return store, nil
	}

	return nil, fmt.Errorf("profile not found")
}

func (uc *StoreUseCase) UpdateProfile(ctx context.Context, in UpdateProfileInput) (*domain.Store, error) {
	var storeID uuid.UUID
	if in.StoreID != "" {
		storeID, _ = uuid.Parse(in.StoreID)
	}

	if storeID == uuid.Nil {
		return nil, fmt.Errorf("StoreID must be provided")
	}

	var dba *string
	if in.DBAName != "" {
		dba = &in.DBAName
	}
	var tax *string
	if in.TaxID != "" {
		tax = &in.TaxID
	}

	store := &domain.Store{
		ID:                storeID,
		LegalName:         strings.TrimSpace(in.BusinessName),
		DBAName:           dba,
		RegisteredAddress: strings.TrimSpace(in.Address),
		Type:              in.BusinessType,
		TaxID:             tax,
		OwnerName:         strings.TrimSpace(in.OwnerName),
		ContactEmail:      strings.ToLower(strings.TrimSpace(in.Email)),
		MobileNumber:      strings.TrimSpace(in.Mobile),
		LogoURL:           in.ProfileImage,
	}

	if storeID != uuid.Nil {
		res, err := uc.repo.UpdateStoreProfileByStoreID(ctx, store)
		if err == nil && res != nil && res.ID != uuid.Nil {
			return res, nil
		}
	}

	return nil, fmt.Errorf("failed to update store profile")
}
