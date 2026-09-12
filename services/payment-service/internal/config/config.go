package config

import (
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	OrderBaseURL  string
	WebhookSecret string
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("PAYMENT_PORT", "8084"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://commerce:commerce123@localhost:5432/commerce?sslmode=disable&search_path=payments"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		OrderBaseURL:  getEnv("ORDER_BASE_URL", "http://localhost:8083"),
		WebhookSecret: getEnv("PAYMENT_WEBHOOK_SECRET", "super-secret-hmac-key-2026"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
