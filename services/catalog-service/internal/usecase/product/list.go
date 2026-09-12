package product

import (
	"context"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
)

type ListProductsInput struct {
	CategoryID string
	Search     string
	MinPrice   int64
	MaxPrice   int64
	ActiveOnly bool
	Page       int
	Limit      int
}

type ListProductsOutput struct {
	Products []*entity.Product
	Total    int
	Page     int
	Limit    int
}

type ListProductsUseCase struct {
	productRepo repository.ProductRepository
}

func NewListProductsUseCase(productRepo repository.ProductRepository) *ListProductsUseCase {
	return &ListProductsUseCase{productRepo: productRepo}
}

func (uc *ListProductsUseCase) Execute(ctx context.Context, input ListProductsInput) (*ListProductsOutput, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}

	offset := (input.Page - 1) * input.Limit

	filter := repository.ProductFilter{
		CategoryID: input.CategoryID,
		Search:     input.Search,
		MinPrice:   input.MinPrice,
		MaxPrice:   input.MaxPrice,
		ActiveOnly: input.ActiveOnly,
		Limit:      input.Limit,
		Offset:     offset,
	}

	products, total, err := uc.productRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListProductsOutput{
		Products: products,
		Total:    total,
		Page:     input.Page,
		Limit:    input.Limit,
	}, nil
}
