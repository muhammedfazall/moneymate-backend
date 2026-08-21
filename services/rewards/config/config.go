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

type RewardsConfig struct {
	PaymentCompletedTopic string `mapstructure:"payment_completed_topic"`
	ConsumerGroup         string `mapstructure:"consumer_group"`
	FakePaymentClient     bool   `mapstructure:"fake_payment_client"`
	PaymentServiceURL     string `mapstructure:"payment_service_url"`
}

type Config struct {
	Env                   string
	Server                ServerConfig  `mapstructure:"server"`
	Rewards               RewardsConfig `mapstructure:"rewards"`
	Database              sharedconfig.DatabaseConfig
	Kafka                 sharedconfig.KafkaConfig
	JWT                   sharedconfig.JWTConfig
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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.Database = sharedconfig.LoadDatabaseConfig(v, "rewards")
	cfg.Kafka = sharedconfig.LoadKafkaConfig(v)
	cfg.JWT = sharedconfig.LoadJWTConfig(v)
	cfg.Env = sharedconfig.Get("ENVIRONMENT", "dev")
	cfg.InternalServiceSecret = sharedconfig.MustGet("INTERNAL_SERVICE_SECRET")
	cfg.Rewards.PaymentServiceURL = sharedconfig.Get("REWARDS_PAYMENT_SERVICE_URL", cfg.Rewards.PaymentServiceURL)

	return &cfg, nil
}
