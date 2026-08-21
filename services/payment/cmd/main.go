package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/moneymate-2026/moneymate-backend/services/payment/config"
	"github.com/moneymate-2026/moneymate-backend/services/payment/internal/app"
	kafkaconsumer "github.com/moneymate-2026/moneymate-backend/services/payment/internal/infra/kafkaconsumer"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to Load Config: %v", err)
	}

	paymentApp, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("Failed to build app: %v", err)
	}
	defer paymentApp.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go paymentApp.KafkaConsumer.Run(ctx, func(ctx context.Context, payload []byte) error {
		return kafkaconsumer.HandleUserRegistered(ctx, paymentApp.WalletUC, payload)
	}) // NEW — starts the Kafka consumer loop

	go paymentApp.OutboxPublisher.Run(ctx) // relays queued events to Kafka

	go func() {
		if err := paymentApp.Run(); err != nil {
			log.Fatalf("App run failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, gracefully shutting down...")
	cancel() // NEW — stops the consumer loop
	paymentApp.Close()
	log.Println("Payment service stopped cleanly")
}
