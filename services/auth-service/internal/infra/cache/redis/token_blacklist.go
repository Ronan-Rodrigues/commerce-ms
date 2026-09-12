package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore implementa a interface service.TokenBlacklist e gerenciamento de refresh tokens
type RedisTokenStore struct {
	client *redis.Client
}

func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{client: client}
}

// Revoke adiciona o JTI do token na blacklist com TTL automático
func (s *RedisTokenStore) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	// Se o TTL for menor que zero, não há necessidade de revogar pois já expirou
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, key, "revoked", ttl).Err()
}

// IsRevoked verifica se o JTI do token consta na blacklist
func (s *RedisTokenStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// StoreRefreshToken armazena o refresh token associado ao ID do usuário com TTL
func (s *RedisTokenStore) StoreRefreshToken(ctx context.Context, refreshToken, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return s.client.Set(ctx, key, userID, ttl).Err()
}

// ValidateRefreshToken verifica e obtém o userID dono do refresh token
func (s *RedisTokenStore) ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	userID, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // não encontrado ou expirado
		}
		return "", err
	}
	return userID, nil
}

// DeleteRefreshToken remove o refresh token (usado para rotação ou logout)
func (s *RedisTokenStore) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return s.client.Del(ctx, key).Err()
}
