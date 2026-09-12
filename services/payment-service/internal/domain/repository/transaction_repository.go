package repository

import (
	"context"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	FindByOrderID(ctx context.Context, orderID string) ([]*entity.Transaction, error)
	GetFinancialSummary(ctx context.Context) (*entity.FinancialSummary, error)
	ListTransactions(ctx context.Context, limit, offset int) ([]*entity.Transaction, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, channel string, event interface{}) error
}

type OrderService interface {
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}

type HMACValidator interface {
	ValidateSignature(payload []byte, signature, secret string) bool
}
