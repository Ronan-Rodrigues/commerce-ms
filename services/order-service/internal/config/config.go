package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	CatalogBaseURL string
	CartTTL        time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:           getEnv("ORDER_PORT", "8083"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://commerce:commerce123@localhost:5432/commerce?sslmode=disable&search_path=orders"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		CatalogBaseURL: getEnv("CATALOG_BASE_URL", "http://localhost:8082"),
		CartTTL:        time.Duration(getEnvInt("CART_TTL_SECONDS", 1800)) * time.Second, // 30 min padrão
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}
