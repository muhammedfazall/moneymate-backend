package sharedconfig

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string

	MaxOpenConns    int
	MinOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
	MaxIdleTime     time.Duration

	DSN string
}


func LoadDatabaseConfig(v *viper.Viper, schema string) DatabaseConfig {
	cfg := DatabaseConfig{
		User:     MustGet("POSTGRES_USER"),
		Password: MustGet("POSTGRES_PASSWORD"),
		Host:     MustGet("POSTGRES_HOST"),
		Port:     Get("POSTGRES_PORT", "5432"),
		Name:     MustGet("POSTGRES_DB"),

		MaxOpenConns:    v.GetInt("database.max_open_conns"),
		MinOpenConns:    v.GetInt("database.min_open_conns"),
		MaxIdleConns:    v.GetInt("database.max_idle_conns"),
		MaxConnLifetime: v.GetDuration("database.max_conn_lifetime"),
		MaxIdleTime:     v.GetDuration("database.max_idle_time"),
	}

	cfg.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, schema,
	)
	return cfg
}