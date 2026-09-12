package product_test

import (
	"context"
	"testing"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/usecase/product"
)

// MockCategoryRepository
type MockCategoryRepository struct {
	categories map[string]*entity.Category
}

func (m *MockCategoryRepository) Create(ctx context.Context, c *entity.Category) error {
	m.categories[c.ID] = c
	return nil
}
func (m *MockCategoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	if c, ok := m.categories[id]; ok {
		return c, nil
	}
	return nil, repository.ErrCategoryNotFound
}
func (m *MockCategoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return nil, nil
}
func (m *MockCategoryRepository) ListAll(ctx context.Context) ([]*entity.Category, error) {
	return nil, nil
}

// MockProductRepository
type MockProductRepository struct {
	products map[string]*entity.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{products: make(map[string]*entity.Product)}
}

func (m *MockProductRepository) Create(ctx context.Context, p *entity.Product) error {
	m.products[p.ID] = p
	return nil
}
func (m *MockProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return nil, repository.ErrProductNotFound
}
func (m *MockProductRepository) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) List(ctx context.Context, filter repository.ProductFilter) ([]*entity.Product, int, error) {
	var list []*entity.Product
	for _, p := range m.products {
		list = append(list, p)
	}
	return list, len(list), nil
}
func (m *MockProductRepository) Update(ctx context.Context, p *entity.Product) error {
	m.products[p.ID] = p
	return nil
}
func (m *MockProductRepository) DeductStockAtomic(ctx context.Context, id string, qty int) error {
	p, ok := m.products[id]
	if !ok {
		return repository.ErrProductNotFound
	}
	return p.DeductStock(qty)
}

func TestCreateProductUseCase_Success(t *testing.T) {
	pRepo := NewMockProductRepository()
	cRepo := &MockCategoryRepository{categories: make(map[string]*entity.Category)}

	cat, _ := entity.NewCategory("cat-1", "Eletrônicos", "Aparelhos em geral")
	_ = cRepo.Create(context.Background(), cat)

	uc := product.NewCreateProductUseCase(pRepo, cRepo)

	out, err := uc.Execute(context.Background(), product.CreateProductInput{
		ID:          "prod-1",
		Name:        "Teclado Mecânico RGB",
		Description: "Switch Blue de alta performance",
		Price:       29990, // R$ 299,90
		Stock:       15,
		SKU:         "TEC-MEC-01",
		CategoryID:  "cat-1",
	})

	if err != nil {
		t.Fatalf("esperava sucesso na criação do produto, obteve: %v", err)
	}

	if out.Slug != "teclado-mecanico-rgb" {
		t.Errorf("slug gerado incorreto: %s", out.Slug)
	}
	if out.Price != 29990 || out.Stock != 15 {
		t.Errorf("dados de preço ou estoque inconsistentes: %+v", out)
	}
}

func TestCreateProductUseCase_InvalidPrice(t *testing.T) {
	pRepo := NewMockProductRepository()
	cRepo := &MockCategoryRepository{categories: make(map[string]*entity.Category)}
	uc := product.NewCreateProductUseCase(pRepo, cRepo)

	_, err := uc.Execute(context.Background(), product.CreateProductInput{
		ID:    "prod-2",
		Name:  "Produto Preço Zero",
		Price: 0, // Preço inválido
		Stock: 5,
		SKU:   "SKU-ZERO",
	})

	if err != entity.ErrInvalidPrice {
		t.Errorf("esperava ErrInvalidPrice, obteve: %v", err)
	}
}

func TestDeductStockUseCase_SuccessAndInsufficient(t *testing.T) {
	pRepo := NewMockProductRepository()

	p, _ := entity.NewProduct("prod-3", "Mouse Gamer", "Mouse 16000 DPI", 15000, 10, "MOU-01", "", "")
	_ = pRepo.Create(context.Background(), p)

	deductUC := product.NewDeductStockUseCase(pRepo)

	// 1. Dedução válida de 4 unidades (restam 6)
	err := deductUC.Execute(context.Background(), product.DeductStockInput{
		ProductID: "prod-3",
		Quantity:  4,
	})
	if err != nil {
		t.Fatalf("erro ao deduzir estoque: %v", err)
	}

	if p.Stock != 6 {
		t.Errorf("esperava estoque restante 6, obteve: %d", p.Stock)
	}

	// 2. Tentativa de deduzir mais do que existe (pedindo 10 quando só tem 6)
	err = deductUC.Execute(context.Background(), product.DeductStockInput{
		ProductID: "prod-3",
		Quantity:  10,
	})
	if err != product.ErrProductOutOfStock {
		t.Errorf("esperava ErrProductOutOfStock, obteve: %v", err)
	}
}
