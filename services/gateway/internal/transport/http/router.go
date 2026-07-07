package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/middlewares"
	"github.com/moneymate-2026/moneymate-backend/gateway/internal/proxy"
)

type Router struct {
	app        *fiber.App
	authClient proxy.AuthClient
}

func NewRouter(auth proxy.AuthClient) *Router {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Prefork:               false, 
		ReduceMemoryUsage:     true,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	return &Router{
		app:        app,
		authClient: auth,
	}
}

func (r *Router) SetupRoutes() {
	api := r.app.Group("/api/v1")

	// Public Route: Health check for load balancers
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Public Routes: Auth endpoints (Login/Register stubs)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", r.handleLogin())
	authGroup.Post("/register", r.handleRegister())

	// --- PROTECTED ROUTES ---
	// We apply the RequireAuth middleware here. Any route added to 'secure' MUST have a valid token.
	secure := api.Group("/secure", middlewares.RequireAuth(r.authClient))
	
	secure.Get("/profile", r.handleProfile())
}

func (r *Router) Listen(addr string) error {
	return r.app.Listen(addr)
}

func (r *Router) Shutdown() error {
	return r.app.Shutdown()
}

// --- Handler Implementations ---

func (r *Router) handleLogin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "login pending gRPC contract"})
	}
}

func (r *Router) handleRegister() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "register pending gRPC contract"})
	}
}

func (r *Router) handleProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the user ID that the middleware securely placed in context
		userID := c.Locals("user_id").(string)
		
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Secure profile accessed",
			"data": fiber.Map{
				"user_id": userID,
			},
		})
	}
}