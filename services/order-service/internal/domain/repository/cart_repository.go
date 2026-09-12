package repository

import (
	"context"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
)

type CartRepository interface {
	Save(ctx context.Context, cart *entity.Cart, ttl time.Duration) error
	Get(ctx context.Context, userID string) (*entity.Cart, error)
	Clear(ctx context.Context, userID string) error
}
