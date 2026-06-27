package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	OTPExpiryMinutes    int
	RateLimitPerMinute  int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	var err error
	cfg.OTPExpiryMinutes, err = strconv.Atoi(getEnv("OTP_EXPIRY_MINUTES", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_EXPIRY_MINUTES: %w", err)
	}

	cfg.RateLimitPerMinute, err = strconv.Atoi(getEnv("RATE_LIMIT_PER_MINUTE", "6"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_PER_MINUTE: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
