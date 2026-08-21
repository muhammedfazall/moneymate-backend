package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moneymate-2026/moneymate-backend/services/rewards/config"
	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/adapter/postgres"
	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/adapter/postgres/repo"
	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/adapter/paymentclient"
	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/domain"
	transporthttp "github.com/moneymate-2026/moneymate-backend/services/rewards/internal/transport/http"
	"github.com/moneymate-2026/moneymate-backend/services/rewards/internal/usecases"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/kafka"
)

type App struct {
	HTTPServer    *fiber.App
	DB            *pgxpool.Pool
	HTTPAddr      string
	KafkaConsumer *kafka.Consumer
	RewardUC      usecases.RewardUsecase
}

func Build(cfg *config.Config) (*App, error) {
	ctx := context.Background()

	pool, err := postgres.ConnectDB(ctx, *cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=rewards",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SslMode,
	)
	if err := postgres.RunMigrations(dsn, cfg.Database.MigrationsPath); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	rewardRepo := repo.NewRewardRepo(pool)

	var paymentClient domain.PaymentClient
	if cfg.Rewards.FakePaymentClient {
		paymentClient = paymentclient.NewFakeClient()
	} else {
		if cfg.Rewards.PaymentServiceURL == "" {
			return nil, fmt.Errorf("rewards.payment_service_url must be set when fake_payment_client is false")
		}
		paymentClient = paymentclient.NewHTTPClient(cfg.Rewards.PaymentServiceURL, cfg.InternalServiceSecret)
	}

	rewardUC := usecases.NewRewardUsecase(rewardRepo, paymentClient)
	rewardHandler := transporthttp.NewRewardHandler(rewardUC)

	server := setupHTTPServer(pool, rewardHandler, cfg.InternalServiceSecret)

	kafkaConsumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:  cfg.Kafka.Brokers,
		Username: cfg.Kafka.Username,
		Password: cfg.Kafka.Password,
		CACert:   cfg.Kafka.CACert,
		Topic:    cfg.Rewards.PaymentCompletedTopic,
		GroupID:  cfg.Rewards.ConsumerGroup,
	})
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	httpAddr := cfg.Server.HTTPAddr
	if port := os.Getenv("PORT"); port != "" {
		httpAddr = port
	}
	if httpAddr == "" {
		httpAddr = "9096"
	}
	if !strings.Contains(httpAddr, ":") {
		httpAddr = "0.0.0.0:" + httpAddr
	}

	return &App{
		HTTPServer:    server,
		DB:            pool,
		HTTPAddr:      httpAddr,
		KafkaConsumer: kafkaConsumer,
		RewardUC:      rewardUC,
	}, nil
}

func setupHTTPServer(pool *pgxpool.Pool, rh *transporthttp.RewardHandler, internalSecret string) *fiber.App {
	server := fiber.New(fiber.Config{AppName: "rewards-service"})

	server.Use(recover.New())
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Internal-Secret"},
	}))

	server.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "rewards"})
	})

	server.Get("/ready", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"status":     "unavailable",
					"dependency": "postgres",
					"error":      err.Error(),
				})
			}
		}

		return c.JSON(fiber.Map{"status": "ready", "service": "rewards"})
	})

	transporthttp.RegisterRoutes(server, rh, internalSecret)
	return server
}

func (a *App) Run() error {
	log.Printf("starting rewards HTTP server on %s", a.HTTPAddr)
	return a.HTTPServer.Listen(a.HTTPAddr)
}

func (a *App) Close() {
	if a.HTTPServer != nil {
		_ = a.HTTPServer.Shutdown()
	}
	if a.DB != nil {
		a.DB.Close()
	}
	if a.KafkaConsumer != nil {
		_ = a.KafkaConsumer.Close()
	}
}

func (a *App) HandleKafkaMessage(ctx context.Context, payload []byte) error {
	log.Printf("rewards kafka message received: %d bytes", len(payload))
	return a.RewardUC.ProcessPaymentCompletedEvent(ctx, payload)
}
