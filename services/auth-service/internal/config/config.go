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
	JWTKeyPath     string // caminho da chave privada RSA
	JWTPublicPath  string // caminho da chave pública RSA
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	Argon2Memory   uint32
	Argon2Iter     uint32
	Argon2Parallel uint8
}

func LoadConfig() *Config {
	return &Config{
		Port:           getEnv("AUTH_PORT", "8081"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://commerce:commerce123@localhost:5432/commerce?sslmode=disable&search_path=auth"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTKeyPath:     getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
		JWTPublicPath:  getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
		JWTAccessTTL:   time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
		JWTRefreshTTL:  time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
		Argon2Memory:   uint32(getEnvInt("ARGON2_MEMORY", 64*1024)),
		Argon2Iter:     uint32(getEnvInt("ARGON2_ITERATIONS", 3)),
		Argon2Parallel: uint8(getEnvInt("ARGON2_PARALLELISM", 2)),
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
