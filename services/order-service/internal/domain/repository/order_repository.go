package repository

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
)

var (
	ErrOrderNotFound = errors.New("pedido não encontrado")
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByIdempotencyKey(ctx context.Context, key string) (*entity.Order, error)
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error)
	UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error
}

type EventPublisher interface {
	Publish(ctx context.Context, channel string, event interface{}) error
}

type CatalogService interface {
	DeductStock(ctx context.Context, productID string, quantity int) error
}
