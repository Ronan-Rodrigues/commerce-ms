package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("CATALOG_PORT", "8082"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://commerce:commerce123@localhost:5432/commerce?sslmode=disable&search_path=catalog"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
