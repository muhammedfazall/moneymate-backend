package sharedconfig

import (
	"time"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Addr          string
	Password      string
	TokenTTL      time.Duration
	OTPTTL        time.Duration
	PINLockoutTTL time.Duration
}

func LoadRedisConfig(v *viper.Viper) RedisConfig {
	return RedisConfig{
		Addr:          MustGet("REDIS_ADDR"),
		Password:      Get("REDIS_PASSWORD", ""),
		TokenTTL:      v.GetDuration("redis.token_ttl"),
		OTPTTL:        v.GetDuration("redis.otp_ttl"),
		PINLockoutTTL: v.GetDuration("redis.pin_lockout_ttl"),
	}
}