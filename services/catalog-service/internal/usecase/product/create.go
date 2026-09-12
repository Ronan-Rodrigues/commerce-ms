package product

import (
	"context"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
)

type CreateProductInput struct {
	ID          string
	Name        string
	Description string
	Price       int64
	Stock       int
	SKU         string
	CategoryID  string
	ImageURL    string
}

type CreateProductOutput struct {
	ID         string
	Name       string
	Slug       string
	Price      int64
	Stock      int
	SKU        string
	CategoryID string
	Active     bool
}

type CreateProductUseCase struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

func NewCreateProductUseCase(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, input CreateProductInput) (*CreateProductOutput, error) {
	// Se foi informada categoria, verifica se existe
	if input.CategoryID != "" {
		_, err := uc.categoryRepo.FindByID(ctx, input.CategoryID)
		if err != nil {
			return nil, repository.ErrCategoryNotFound
		}
	}

	product, err := entity.NewProduct(
		input.ID,
		input.Name,
		input.Description,
		input.Price,
		input.Stock,
		input.SKU,
		input.CategoryID,
		input.ImageURL,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return &CreateProductOutput{
		ID:         product.ID,
		Name:       product.Name,
		Slug:       product.Slug,
		Price:      product.Price,
		Stock:      product.Stock,
		SKU:        product.SKU,
		CategoryID: product.CategoryID,
		Active:     product.Active,
	}, nil
}
