package repository

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
)

var (
	ErrCategoryNotFound      = errors.New("categoria não encontrada")
	ErrCategoryAlreadyExists = errors.New("já existe uma categoria com este nome ou slug")
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	ListAll(ctx context.Context) ([]*entity.Category, error)
}
