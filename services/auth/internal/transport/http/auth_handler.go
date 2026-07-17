package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/moneymate-2026/moneymate-backend/auth/internal/domain"
	usecase "github.com/moneymate-2026/moneymate-backend/auth/internal/usecases"
	response "github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

var validate = validator.New()

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	otpUsecase  usecase.OTPUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase, otpUsecase usecase.OTPUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		otpUsecase:  otpUsecase,
	}
}


func (h *AuthHandler) Register(accountType domain.AccountType) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req registerRequest
		if err := c.Bind().Body(&req); err != nil {
			return response.BadRequest(c, nil, "invalid request body")
		}
		if err := validate.Struct(req); err != nil {
			return response.BadRequest(c, formatValidationErrors(err), "validation failed")
		}

		ucReq := usecase.RegisterRequest{
			Email:       req.Email,
			Phone:       req.Phone,
			FullName:    req.FullName,
			Password:    req.Password,
			AccountType: accountType,
		}

		resp, err := h.authUsecase.Register(c.Context(), ucReq)
		if err != nil {
			return handleError(c, err)
		}
		return response.Created(c, "account created successfully", resp)
	}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c, formatValidationErrors(err), "validation failed")
	}

	deviceID := c.Get("X-Device-Id")
	if deviceID == "" {
		return response.BadRequest(c, nil, "X-Device-Id header is required")
	}

	ucReq := usecase.LoginRequest{
		Identifier: req.Email,
		Password:   req.Password,
		DeviceID:   deviceID,
		UserAgent:  c.Get("User-Agent"),
		IPAddress:  c.IP(),
	}

	resp, err := h.authUsecase.Login(c.Context(), ucReq)
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "login successful", resp)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return response.Unauthorized(c, "authentication required")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "invalid user session")
	}

	var req logoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	ucReq := usecase.LogoutRequest{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
		AllDevices:   req.AllDevices,
	}

	if err := h.authUsecase.Logout(c.Context(), ucReq); err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "logged out successfully", nil)
}

func (h *AuthHandler) SendRegistrationOTP(c fiber.Ctx) error {
	var req sendRegistrationOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c, formatValidationErrors(err), "validation failed")
	}

	resp, err := h.otpUsecase.SendRegistrationOTP(c.Context(), usecase.SendRegistrationOTPRequest{
		Email: req.Email,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "verification code sent", resp)
}

func (h *AuthHandler) VerifyRegistrationOTP(c fiber.Ctx) error {
	var req verifyRegistrationOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return response.BadRequest(c, formatValidationErrors(err), "validation failed")
	}

	resp, err := h.otpUsecase.VerifyRegistrationOTP(c.Context(), usecase.VerifyRegistrationOTPRequest{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		return handleError(c, err)
	}
	return response.OK(c, "email verified", resp)
}