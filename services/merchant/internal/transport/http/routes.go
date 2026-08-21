package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/moneymate-2026/moneymate-backend/services/merchant/internal/transport/http/middleware"
)

// RegisterRoutes wires all HTTP endpoints for merchant registration, campaigns, rewards center, subscription plans, KYC compliance, and Wallet.
func RegisterRoutes(router fiber.Router, h *MerchantHandler, ch *CampaignHandler, rh *RewardHandler, sh *SubscriptionHandler, kh *KYCHandler, dh *DashboardHandler, wh *WalletHandler, eh *EarningsHandler, authMiddleware fiber.Handler, internalSecret string) {
	merchant := router.Group("/merchant")

	merchant.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"service": "merchant",
		})
	})

	internal := router.Group("/internal/merchant", middleware.RequireInternalSecret(internalSecret))
	internal.Get("/stores/:id/profile", h.GetInternalProfile)

	merchant.Post("/register", h.RegisterStore)
	merchant.Post("/login", h.LoginStore)

	// Public routes that do not require authentication
	merchant.Get("/public/campaigns", ch.GetAllPublicCampaigns)
	merchant.Get("/public/subscriptions/plans", sh.GetPublicPlans)

	// Apply auth middleware if provided
	if authMiddleware != nil {
		merchant.Use(authMiddleware)
	}

	merchant.Get("/status/:store_id", h.GetStore)
	merchant.Get("/pending", h.GetPendingStores)

	// Profile & Settings routes
	merchant.Get("/profile", h.GetProfile)
	merchant.Get("/:store_id/profile", h.GetProfile)
	merchant.Put("/profile", h.UpdateProfile)
	merchant.Put("/:store_id/profile", h.UpdateProfile)
	merchant.Post("/profile", h.UpdateProfile)
	merchant.Post("/:store_id/profile", h.UpdateProfile)

	// Dashboard routes
	if dh != nil {
		merchant.Get("/dashboard", dh.GetDashboard)
		merchant.Get("/:store_id/dashboard", dh.GetDashboard)
	}

	// Campaign routes
	merchant.Post("/campaigns", ch.CreateCampaign)
	merchant.Post("/:store_id/campaigns", ch.CreateCampaign)
	merchant.Get("/campaigns", ch.GetCampaigns)
	merchant.Get("/:store_id/campaigns", ch.GetCampaigns)
	merchant.Put("/campaigns/:id/status", ch.UpdateCampaignStatus)
	merchant.Put("/:store_id/campaigns/:id/status", ch.UpdateCampaignStatus)

	// Rewards Center routes
	if rh != nil {
		merchant.Get("/rewards/summary", rh.GetRewardSummary)
		merchant.Get("/:store_id/rewards/summary", rh.GetRewardSummary)
		merchant.Get("/rewards/history", rh.GetRewardHistory)
		merchant.Get("/:store_id/rewards/history", rh.GetRewardHistory)
		merchant.Post("/rewards/redeem", rh.RedeemRewards)
		merchant.Post("/:store_id/rewards/redeem", rh.RedeemRewards)
	}

	// Subscription Plans & Billing routes
	if sh != nil {
		merchant.Get("/subscriptions/plans", sh.GetSubscriptionPlans)
		merchant.Get("/:store_id/subscriptions/plans", sh.GetSubscriptionPlans)
		merchant.Get("/subscriptions/current", sh.GetCurrentSubscription)
		merchant.Get("/:store_id/subscriptions/current", sh.GetCurrentSubscription)
		merchant.Post("/subscriptions/change", sh.ChangeSubscriptionPlan)
		merchant.Post("/:store_id/subscriptions/change", sh.ChangeSubscriptionPlan)
		merchant.Post("/subscriptions/upgrade/initiate", sh.InitiateUpgrade)
		merchant.Post("/:store_id/subscriptions/upgrade/initiate", sh.InitiateUpgrade)
		merchant.Post("/subscriptions/upgrade/verify", sh.VerifyUpgrade)
		merchant.Post("/:store_id/subscriptions/upgrade/verify", sh.VerifyUpgrade)
	}

	// Wallet routes
	if wh != nil {
		merchant.Get("/wallet", wh.GetWallet)
		merchant.Get("/:store_id/wallet", wh.GetWallet)
	}

	// Earnings routes
	if eh != nil {
		merchant.Get("/earnings", eh.GetEarnings)
		merchant.Get("/:store_id/earnings", eh.GetEarnings)
		merchant.Post("/earnings/payouts", eh.RequestPayout)
		merchant.Post("/:store_id/earnings/payouts", eh.RequestPayout)
	}

	// KYC Verification & Compliance routes
	if kh != nil {
		merchant.Get("/kyc", kh.GetKYCStatus)
		merchant.Get("/:store_id/kyc", kh.GetKYCStatus)
		merchant.Get("/kyc/status", kh.GetKYCStatus)
		merchant.Get("/:store_id/kyc/status", kh.GetKYCStatus)
		merchant.Put("/kyc", kh.UpdateKYCDocuments)
		merchant.Put("/:store_id/kyc", kh.UpdateKYCDocuments)
		merchant.Post("/kyc/update", kh.UpdateKYCDocuments)
		merchant.Post("/:store_id/kyc/update", kh.UpdateKYCDocuments)
	}
}

// RegisterWebSocketRoutes wires the websocket endpoint for merchant live updates
func RegisterWebSocketRoutes(router fiber.Router, handler fiber.Handler) {
	router.Get("/ws", handler)
}

// RegisterAdminRoutes wires all HTTP endpoints for platform administrators to perform CRUD and governance operations.
func RegisterAdminRoutes(router fiber.Router, ah *AdminHandler, authMiddleware fiber.Handler) {
	if ah == nil {
		return
	}

	setupAdminGroup := func(grp fiber.Router) {
		if authMiddleware != nil {
			grp.Use(authMiddleware)
		}

		// Stores / Merchants
		grp.Get("/stores", ah.GetAllStores)
		grp.Get("/merchants", ah.GetAllStores)
		grp.Get("/stores/:id", ah.GetStoreByID)
		grp.Get("/merchants/:id", ah.GetStoreByID)
		grp.Put("/stores/:id/status", ah.UpdateStoreStatus)
		grp.Put("/merchants/:id/status", ah.UpdateStoreStatus)
		grp.Delete("/stores/:id", ah.DeleteStore)
		grp.Delete("/merchants/:id", ah.DeleteStore)

		// Campaigns
		grp.Get("/campaigns", ah.GetAllCampaigns)
		grp.Get("/stores/:store_id/campaigns", ah.GetCampaignsByStoreID)
		grp.Get("/merchants/:store_id/campaigns", ah.GetCampaignsByStoreID)
		grp.Post("/stores/:store_id/campaigns", ah.CreateCampaign)
		grp.Post("/merchants/:store_id/campaigns", ah.CreateCampaign)
		grp.Put("/campaigns/:id/status", ah.UpdateCampaignStatus)
		grp.Delete("/campaigns/:id", ah.DeleteCampaign)

		// KYC Verification
		grp.Get("/kyc", ah.GetAllKYCDocuments)
		grp.Get("/kyc/:store_id", ah.GetKYCByStoreID)
		grp.Get("/stores/:store_id/kyc", ah.GetKYCByStoreID)
		grp.Get("/merchants/:store_id/kyc", ah.GetKYCByStoreID)
		grp.Put("/kyc/:store_id/verify", ah.VerifyKYCDocument)
		grp.Put("/stores/:store_id/kyc/verify", ah.VerifyKYCDocument)
		grp.Put("/merchants/:store_id/kyc/verify", ah.VerifyKYCDocument)

		// Rewards
		grp.Get("/rewards/history", ah.GetAllRewardTransactions)
		grp.Get("/rewards/summary", ah.GetPlatformRewardSummary)

		// Subscriptions
		grp.Get("/subscriptions", ah.GetAllSubscriptions)
		grp.Put("/stores/:store_id/subscription", ah.UpdateStoreSubscriptionPlan)
		grp.Put("/merchants/:store_id/subscription", ah.UpdateStoreSubscriptionPlan)

		// Dashboard
		grp.Get("/dashboard/stats", ah.GetAdminDashboardStats)
	}


	setupAdminGroup(router.Group("/admin"))
	setupAdminGroup(router.Group("/merchant/admin"))
	setupAdminGroup(router.Group("/admin/merchant"))
}
