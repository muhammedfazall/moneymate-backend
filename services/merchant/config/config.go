package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	sharedconfig "github.com/moneymate-2026/moneymate-backend/shared/config"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	HTTPAddr string `mapstructure:"http_addr"`
}

type RazorpayConfig struct {
	KeyID         string `mapstructure:"key_id"`
	KeySecret     string `mapstructure:"key_secret"`
	WebhookSecret string `mapstructure:"webhook_secret"`
}

type Config struct {
	Env      string
	Server   ServerConfig `mapstructure:"server"`
	Database sharedconfig.DatabaseConfig
	Razorpay RazorpayConfig
	Kafka    sharedconfig.KafkaConfig
	InternalServiceSecret string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = "./config/config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(yamlPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Ignore if file doesn't exist, rely on env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: failed to read config file: %v\n", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.Database = sharedconfig.LoadDatabaseConfig(v, "merchant")
	cfg.Env = sharedconfig.Get("ENVIRONMENT", "dev")

	cfg.Razorpay.KeyID = sharedconfig.Get("RAZORPAY_KEY_ID", "")
	cfg.Razorpay.KeySecret = sharedconfig.Get("RAZORPAY_KEY_SECRET", "")
	cfg.Razorpay.WebhookSecret = sharedconfig.Get("RAZORPAY_WEBHOOK_SECRET", "")
	cfg.Kafka = sharedconfig.LoadKafkaConfig(v)
	cfg.InternalServiceSecret = sharedconfig.MustGet("INTERNAL_SERVICE_SECRET")

	return &cfg, nil
}
