package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

type RedisCartRepository struct {
	client *redis.Client
}

func NewRedisCartRepository(client *redis.Client) *RedisCartRepository {
	return &RedisCartRepository{client: client}
}

func (r *RedisCartRepository) Save(ctx context.Context, cart *entity.Cart, ttl time.Duration) error {
	key := fmt.Sprintf("cart:%s", cart.UserID)
	snapshotKey := fmt.Sprintf("cart_snapshot:%s", cart.UserID)

	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	// Chave principal com TTL de 30min (expira para abandono)
	pipe.Set(ctx, key, data, ttl)
	// Snapshot preservado por 24h para consulta dos itens quando o abandono ocorrer
	pipe.Set(ctx, snapshotKey, data, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisCartRepository) Get(ctx context.Context, userID string) (*entity.Cart, error) {
	key := fmt.Sprintf("cart:%s", userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // carrinho não existe ou expirou
		}
		return nil, err
	}

	var c entity.Cart
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *RedisCartRepository) Clear(ctx context.Context, userID string) error {
	key := fmt.Sprintf("cart:%s", userID)
	snapshotKey := fmt.Sprintf("cart_snapshot:%s", userID)
	return r.client.Del(ctx, key, snapshotKey).Err()
}

// GetSnapshot recupera os itens do carrinho que expirou
func (r *RedisCartRepository) GetSnapshot(ctx context.Context, userID string) (*entity.Cart, error) {
	snapshotKey := fmt.Sprintf("cart_snapshot:%s", userID)
	data, err := r.client.Get(ctx, snapshotKey).Bytes()
	if err != nil {
		return nil, err
	}

	var c entity.Cart
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
