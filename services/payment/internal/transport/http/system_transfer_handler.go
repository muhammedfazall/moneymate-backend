package http

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/usecases"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/money"
	response "github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

type SystemTransferHandler struct {
	systemTransfers usecases.SystemTransferUsecase
}

func NewSystemTransferHandler(systemTransfers usecases.SystemTransferUsecase) *SystemTransferHandler {
	return &SystemTransferHandler{systemTransfers: systemTransfers}
}

type systemTransferRequest struct {
	FromAccountID  string `json:"from_account_id" validate:"required,uuid"`
	ToAccountID    string `json:"to_account_id" validate:"required,uuid"`
	AmountPaise    int64  `json:"amount_paise" validate:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
	Description    string `json:"description"`
}

func (h *SystemTransferHandler) Transfer(c fiber.Ctx) error {
	var req systemTransferRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c,formatValidationErrors(err), "validation failed")
	}

	fromID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid from_account_id")
	}
	toID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid to_account_id")
	}

	result, err := h.systemTransfers.Transfer(c.Context(), usecases.SystemTransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toID,
		AmountPaise:    req.AmountPaise,
		IdempotencyKey: req.IdempotencyKey,
		Description:    req.Description,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "system transfer completed", transferResponse{
		Transaction: toTransactionResponse(result.Transaction),
		FromBalance: money.FormatPaise(result.FromBalance),
		ToBalance:   money.FormatPaise(result.ToBalance),
	})
}

func formatValidationErrors(err error) []string {
	var out []string
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			out = append(out, fmt.Sprintf("%s: %s", fe.Field(), fe.Tag()))
		}
	}
	return out
}