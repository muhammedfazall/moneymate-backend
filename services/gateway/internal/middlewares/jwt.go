package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
)

// RequireAuth forces the request to possess a valid token verified by the Auth Service.
func RequireAuth(authClient proxy.AuthClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "missing authorization header",
			})
		}

		// Fast, zero-allocation prefix check
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid token format, expected Bearer",
			})
		}

		// Extract the actual token string
		token := authHeader[len(prefix):]

		// Call the Auth proxy (currently your MockAuthClient)
		userID, err := authClient.VerifyToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid or expired token",
			})
		}

		// Inject the verified User ID into Fiber's localized memory context
		c.Locals("user_id", userID)

		// Proceed to the downstream handler
		return c.Next()
	}
}