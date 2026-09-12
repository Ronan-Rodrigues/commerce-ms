package config

import (
	"os"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	N8NBaseURL    string
	N8NSecret     string
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("NOTIFY_PORT", "8085"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		N8NBaseURL:    getEnv("N8N_BASE_URL", "http://localhost:5678"),
		N8NSecret:     getEnv("N8N_WEBHOOK_SECRET", "n8n-secret-token-2026"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
