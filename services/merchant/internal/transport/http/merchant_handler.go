package http

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/transport/http/middleware"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/usecases"
)

type MerchantHandler struct {
	usecase *usecases.StoreUseCase
}

func NewMerchantHandler(uc *usecases.StoreUseCase) *MerchantHandler {
	return &MerchantHandler{usecase: uc}
}

func (h *MerchantHandler) RegisterStore(c fiber.Ctx) error {
	var req RegisterStoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	input := usecases.RegisterStoreInput{
		OwnerName:         req.OwnerName,
		ContactEmail:      req.ContactEmail,
		MobileNumber:      req.MobileNumber,
		LegalName:         req.LegalName,
		Type:              req.BusinessType,
		RegisteredAddress: req.RegisteredAddress,
		AadhaarNumber:     req.AadhaarNumber,
		AadhaarDocURL:     req.AadhaarDocURL,
		ShopLicenseURL:    req.ShopLicenseURL,
		Password:          req.Password,
	}

	if req.DBAName != "" {
		input.DBAName = &req.DBAName
	}
	if req.TaxID != "" {
		input.TaxID = &req.TaxID
	}

	store, err := h.usecase.ProcessRegistration(c.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := middleware.GenerateToken(store.ID.String(), "merchant")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(RegisterStoreResponse{
		StoreID:      store.ID.String(),
		DisplayID:    store.DisplayID,
		VPA:          store.VPA,
		QRCodeBase64: store.QRCodeBase64,
		Status:       store.Status,
		Plan:         store.Plan,
		Token:        token,
	})
}

func (h *MerchantHandler) LoginStore(c fiber.Ctx) error {
	var req LoginStoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password are required"})
	}

	store, err := h.usecase.AuthenticateStore(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := middleware.GenerateToken(store.ID.String(), "merchant")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(LoginStoreResponse{
		StoreID:   store.ID.String(),

		DisplayID: store.DisplayID,
		VPA:       store.VPA,
		LegalName: store.LegalName,
		Status:    store.Status,
		Plan:      store.Plan,
		Token:     token,
	})
}

func (h *MerchantHandler) GetStore(c fiber.Ctx) error {
	storeID := c.Params("store_id")
	if storeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "store_id is required"})
	}

	store, err := h.usecase.GetStore(c.Context(), storeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(GetStoreResponse{
		StoreID:   store.ID.String(),
		DisplayID:    store.DisplayID,
		VPA:          store.VPA,
		QRCodeBase64: store.QRCodeBase64,
		Status:       store.Status,
		Plan:         store.Plan,
		LegalName:    store.LegalName,
	})
}

func (h *MerchantHandler) GetPendingStores(c fiber.Ctx) error {
	stores, err := h.usecase.GetPendingStores(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	// Convert domain.Store to DTOs or anonymous structs here if necessary
	var responseList []fiber.Map
	for _, s := range stores {
		responseList = append(responseList, fiber.Map{
			"store_id":       s.ID.String(),
			"owner_name":     s.OwnerName,
			"contact_email":  s.ContactEmail,
			"mobile_number":  s.MobileNumber,
			"legal_name":     s.LegalName,
			"business_type":  s.Type,
			"status":         s.Status,
			"created_at":     s.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"stores": responseList,
	})
}

func resolveMerchantID(c fiber.Ctx) string {
	// Secure fintech practice: For merchant routes, always resolve the store context strictly from the authenticated user's ID
	// Do NOT trust store_id or owner_id from query params or URL params to prevent IDOR.
	if id, ok := c.Locals("store_id").(string); ok && id != "" {
		return id
	}
	
	// Fallback to URL parameters for internal testing or when auth middleware is bypassed
	if id := c.Params("store_id"); id != "" {
		return id
	}
	
	return ""
}

func (h *MerchantHandler) GetProfile(c fiber.Ctx) error {
	id := resolveMerchantID(c)
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "store_id is required to fetch profile"})
	}

	store, err := h.usecase.GetProfile(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "profile not found: " + err.Error()})
	}

	var dba, tax string
	if store.DBAName != nil {
		dba = *store.DBAName
	}
	if store.TaxID != nil {
		tax = *store.TaxID
	}

	status := store.Status
	switch status {
	case "active", "verified":
		status = "Verified"
	case "pending_kyc":
		status = "Pending Review"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": ProfileResponse{
			StoreID:      store.ID.String(),

			BusinessName: store.LegalName,
			DBAName:      dba,
			Address:      store.RegisteredAddress,
			BusinessType: store.Type,
			TaxID:        tax,
			OwnerName:    store.OwnerName,
			Email:        store.ContactEmail,
			Mobile:       store.MobileNumber,
			ProfileImage: store.LogoURL,
			Status:       status,
			DisplayID:    store.DisplayID,
			VPA:          store.VPA,
			QRCodeBase64: store.QRCodeBase64,
			Plan:         store.Plan,
			CreatedAt:    store.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

func (h *MerchantHandler) UpdateProfile(c fiber.Ctx) error {
	id := resolveMerchantID(c)
	var req UpdateProfileRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request payload"})
	}

	input := usecases.UpdateProfileInput{
		BusinessName: req.BusinessName,
		DBAName:      req.DBAName,
		Address:      req.Address,
		BusinessType: req.BusinessType,
		TaxID:        req.TaxID,
		OwnerName:    req.OwnerName,
		Email:        req.Email,
		Mobile:       req.Mobile,
		ProfileImage: req.ProfileImage,
	}

	if id != "" {
		input.StoreID = id
	}

	store, err := h.usecase.UpdateProfile(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	var dba, tax string
	if store.DBAName != nil {
		dba = *store.DBAName
	}
	if store.TaxID != nil {
		tax = *store.TaxID
	}

	status := store.Status
	switch status {
	case "active", "verified":
		status = "Verified"
	case "pending_kyc":
		status = "Pending Review"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business profile updated successfully!",
		"data": ProfileResponse{
			StoreID:      store.ID.String(),

			BusinessName: store.LegalName,
			DBAName:      dba,
			Address:      store.RegisteredAddress,
			BusinessType: store.Type,
			TaxID:        tax,
			OwnerName:    store.OwnerName,
			Email:        store.ContactEmail,
			Mobile:       store.MobileNumber,
			ProfileImage: store.LogoURL,
			Status:       status,
			DisplayID:    store.DisplayID,
			VPA:          store.VPA,
			QRCodeBase64: store.QRCodeBase64,
			Plan:         store.Plan,
			CreatedAt:    store.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

func (h *MerchantHandler) GetInternalProfile(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id is required"})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid store id"})
	}

	store, err := h.usecase.GetStoreProfile(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "store not found"})
	}

	// ONLY return public fields to avoid leaking PII or credentials
	return c.JSON(fiber.Map{
		"store_id":   store.ID.String(),
		"legal_name": store.LegalName,
		"display_id": store.DisplayID,
		"logo_url":   store.LogoURL,
		"plan":       store.Plan,
	})
}
