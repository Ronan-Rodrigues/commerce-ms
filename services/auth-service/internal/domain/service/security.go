package service

import (
	"context"
	"time"
)

// PasswordHasher define o contrato para hashing de senhas (ex: Argon2id)
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password, hash string) bool
}

// TokenPayload representa os dados contidos dentro do JWT
type TokenPayload struct {
	UserID string
	Email  string
	Role   string
}

// TokenService define o contrato para geração e validação de tokens JWT
type TokenService interface {
	GenerateAccessToken(payload TokenPayload, ttl time.Duration) (string, error)
	GenerateRefreshToken() (string, error)
}

// TokenBlacklist define o contrato para revogação de tokens no cache (ex: Redis)
type TokenBlacklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

