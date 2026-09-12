package repository

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
)

var (
	ErrProductNotFound      = errors.New("produto não encontrado")
	ErrProductAlreadyExists = errors.New("já existe um produto com este SKU")
)

type ProductFilter struct {
	CategoryID string
	Search     string
	MinPrice   int64
	MaxPrice   int64
	ActiveOnly bool
	Limit      int
	Offset     int
}

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
	List(ctx context.Context, filter ProductFilter) ([]*entity.Product, int, error)
	Update(ctx context.Context, product *entity.Product) error
	// DeductStockAtomic deduz quantidade no estoque de forma atômica no banco (proteção contra concorrência)
	DeductStockAtomic(ctx context.Context, id string, quantity int) error
}
