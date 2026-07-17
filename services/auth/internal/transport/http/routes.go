package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/moneymate-2026/moneymate-backend/auth/internal/domain"
)

func RegisterRoutes(router fiber.Router, h *AuthHandler, authMiddleware fiber.Handler) {
	userAuth := router.Group("/user/auth")
	userAuth.Post("/register", h.Register(domain.AccountTypeUser))
	userAuth.Post("/login", h.Login)
	userAuth.Post("/logout", authMiddleware, h.Logout)
	userAuth.Post("/otp/send", h.SendRegistrationOTP)
	userAuth.Post("/otp/verify", h.VerifyRegistrationOTP)

	merchantAuth := router.Group("/merchant/auth")
	merchantAuth.Post("/register", h.Register(domain.AccountTypeMerchant))
	merchantAuth.Post("/login", h.Login)
	merchantAuth.Post("/logout", authMiddleware, h.Logout)
	merchantAuth.Post("/otp/send", h.SendRegistrationOTP)
	merchantAuth.Post("/otp/verify", h.VerifyRegistrationOTP)
}
