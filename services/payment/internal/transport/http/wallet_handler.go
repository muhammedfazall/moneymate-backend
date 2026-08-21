package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/domain"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/usecases"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/money"
	response "github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

type WalletHandler struct {
	wallets usecases.WalletUsecase
}

func NewWalletHandler(wallets usecases.WalletUsecase) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}
func (h *WalletHandler) CreateWallet(c fiber.Ctx) error {
	userID := userIDFromLocals(c)
	if userID == "" {
		return response.Unauthorized(c, "authentication required")
	}
	return response.BadRequest(c, nil, "wallets are created automatically at registration")
}

func (h *WalletHandler) CreateWalletInternal(c fiber.Ctx) error {
	var req createWalletInternalRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c, nil, "validation failed")
	}

	acc, err := h.wallets.CreateWallet(c.Context(), req.UserID, req.Handle)
	if err != nil {
		return handleError(c, err)
	}
	return response.Created(c, "wallet created", toWalletResponse(acc))
}

type creditRewardRequest struct {
	PayoutID    string `json:"payout_id"`
	AmountPaise int64  `json:"amount_paise" validate:"required,gt=0"`
}

// CreditReward is an internal endpoint used by the rewards service to pay
// cashback into a user's wallet. Idempotent on payout_id.
func (h *WalletHandler) CreditReward(c fiber.Ctx) error {
	accountID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, nil, "invalid account id")
	}
	var req creditRewardRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c, nil, "validation failed")
	}
	payoutID, err := uuid.Parse(req.PayoutID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid payout id")
	}

	txID, err := h.wallets.CreditReward(c.Context(), accountID, payoutID, req.AmountPaise)
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "reward credited", fiber.Map{
		"transaction_id": txID.String(),
		"status":         "completed",
	})
}


func (h *WalletHandler) GetMyWallet(c fiber.Ctx) error {
	userID := userIDFromLocals(c)
	if userID == "" {
		return response.Unauthorized(c, "authentication required")
	}
	result, err := h.wallets.GetWalletWithTotal(c.Context(), userID)
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "wallet found", fiber.Map{
		"wallet":        toWalletResponse(result.Wallet),
		"total_balance": money.FormatPaise(result.TotalBalance),
	})
}

func (h *WalletHandler) ListMyAccounts(c fiber.Ctx) error {
	userID := userIDFromLocals(c)
	if userID == "" {
		return response.Unauthorized(c, "authentication required")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid user id")
	}
	accounts, err := h.wallets.ListAccounts(c.Context(), uid)
	if err != nil {
		return handleError(c, err)
	}
	out := make([]walletResponse, len(accounts))
	for i, a := range accounts {
		out[i] = toWalletResponse(a)
	}
	return response.OK(c, "accounts listed", out)
}

func (h *WalletHandler) GetWalletByID(c fiber.Ctx) error {
	acc, err := h.wallets.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return handleError(c, err)
	}
	// Only the owner can view their own wallet by ID either.
	userID := userIDFromLocals(c)
	if acc.UserID == nil || acc.UserID.String() != userID {
		return response.Forbidden(c, nil, "you do not have access to this wallet")
	}
	return response.OK(c, "wallet found", toWalletResponse(acc))
}

func toWalletResponse(a *domain.Account) walletResponse {
	var userID string
	if a.UserID != nil {
		userID = a.UserID.String()
	}
	return walletResponse{
		ID:       a.ID.String(),
		UserID:   userID,
		Type:     string(a.Type),
		Currency: a.Currency,
		Balance:  money.FormatPaise(a.Balance),
	}
}
