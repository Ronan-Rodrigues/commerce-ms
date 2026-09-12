package product

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
)

var (
	ErrProductOutOfStock = errors.New("quantidade em estoque indisponível para este produto")
)

type DeductStockInput struct {
	ProductID string
	Quantity  int
}

type DeductStockUseCase struct {
	productRepo repository.ProductRepository
}

func NewDeductStockUseCase(productRepo repository.ProductRepository) *DeductStockUseCase {
	return &DeductStockUseCase{productRepo: productRepo}
}

func (uc *DeductStockUseCase) Execute(ctx context.Context, input DeductStockInput) error {
	if input.Quantity <= 0 {
		return entity.ErrInvalidQuantity
	}

	product, err := uc.productRepo.FindByID(ctx, input.ProductID)
	if err != nil {
		return repository.ErrProductNotFound
	}

	if product.Stock < input.Quantity {
		return ErrProductOutOfStock
	}

	// Executa dedução atômica no banco de dados
	return uc.productRepo.DeductStockAtomic(ctx, input.ProductID, input.Quantity)
}
