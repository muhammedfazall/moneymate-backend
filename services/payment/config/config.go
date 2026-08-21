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

type Config struct {
	Env                   string
	Server                ServerConfig `mapstructure:"server"`
	Database              sharedconfig.DatabaseConfig
	JWT                   sharedconfig.JWTConfig
	Kafka    sharedconfig.KafkaConfig
	Razorpay sharedconfig.RazorpayConfig
	AuthServiceURL        string
	MerchantServiceURL    string
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
			fmt.Printf("Warning: failed to read config file: %v\n", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.Database = sharedconfig.LoadDatabaseConfig(v, "payment")
	cfg.JWT = sharedconfig.LoadJWTConfig(v)
	cfg.Razorpay = sharedconfig.LoadRazorpayConfig(v)
	cfg.Env = sharedconfig.Get("ENVIRONMENT", "dev")
	cfg.Kafka = sharedconfig.LoadKafkaConfig(v)
	cfg.AuthServiceURL = sharedconfig.Get("AUTH_SERVICE_URL", "http://auth:8081")
	cfg.MerchantServiceURL = sharedconfig.Get("MERCHANT_SERVICE_URL", "http://merchant:8082")
	cfg.InternalServiceSecret = sharedconfig.MustGet("INTERNAL_SERVICE_SECRET")
	return &cfg, nil
}
