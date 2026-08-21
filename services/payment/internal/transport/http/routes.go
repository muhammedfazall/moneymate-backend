package http

import (
	"github.com/gofiber/fiber/v3"

	authclient "github.com/moneymate-2026/moneymate-backend/services/payment/internal/adapter/authClient"
	sharedjwt "github.com/moneymate-2026/moneymate-backend/shared/pkg/jwt"
)

func RegisterRoutes(router fiber.Router, wh *WalletHandler, th *TransferHandler, dh *DepositHandler, wdh *WithdrawalHandler, ch *CategoryHandler, jwtCfg sharedjwt.Config, authClient *authclient.Client, internalSecret string) {	pay := router.Group("/payment", RequireUserID(jwtCfg))

	pay.Get("/wallets/me", RequireTransactionToken(authClient), wh.GetMyWallet)
	pay.Get("/wallets/:id", wh.GetWalletByID)

	pay.Post("/transfers", RequireTransactionToken(authClient), th.Transfer)
	pay.Get("/transactions/:id", th.GetTransaction)
	
	pay.Post("/categories", ch.Create)
	pay.Get("/categories", ch.List)
	pay.Put("/categories/:id", ch.Update)
	pay.Delete("/categories/:id", ch.Delete)

	pay.Post("/deposits", RequireTransactionToken(authClient), dh.Initiate)
	pay.Post("/deposits/confirm", dh.Confirm)
	pay.Get("/deposits", dh.List)

	pay.Post("/withdrawals", RequireTransactionToken(authClient), wdh.Request)
	pay.Get("/withdrawals", wdh.List)

	internal := router.Group("/internal", RequireInternalSecret(internalSecret))
	internal.Post("/payment/wallets", wh.CreateWalletInternal)
	internal.Post("/payment/wallets/:id/credit-reward", wh.CreditReward)
}
