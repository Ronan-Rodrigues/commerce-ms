package order_test

import (
	"context"
	"testing"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/usecase/order"
)

// MockOrderRepository
type MockOrderRepository struct {
	orders   map[string]*entity.Order
	byKey    map[string]*entity.Order
	byUserID map[string][]*entity.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders:   make(map[string]*entity.Order),
		byKey:    make(map[string]*entity.Order),
		byUserID: make(map[string][]*entity.Order),
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, o *entity.Order) error {
	m.orders[o.ID] = o
	if o.IdempotencyKey != "" {
		m.byKey[o.IdempotencyKey] = o
	}
	m.byUserID[o.UserID] = append(m.byUserID[o.UserID], o)
	return nil
}

func (m *MockOrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*entity.Order, error) {
	if o, ok := m.byKey[key]; ok {
		return o, nil
	}
	return nil, repository.ErrOrderNotFound
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	if o, ok := m.orders[id]; ok {
		return o, nil
	}
	return nil, repository.ErrOrderNotFound
}

func (m *MockOrderRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	return m.byUserID[userID], nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	if o, ok := m.orders[id]; ok {
		return o.TransitionTo(status)
	}
	return repository.ErrOrderNotFound
}

// MockCartRepository
type MockCartRepository struct {
	carts map[string]*entity.Cart
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{carts: make(map[string]*entity.Cart)}
}

func (m *MockCartRepository) Save(ctx context.Context, c *entity.Cart, ttl time.Duration) error {
	m.carts[c.UserID] = c
	return nil
}

func (m *MockCartRepository) Get(ctx context.Context, userID string) (*entity.Cart, error) {
	return m.carts[userID], nil
}

func (m *MockCartRepository) Clear(ctx context.Context, userID string) error {
	delete(m.carts, userID)
	return nil
}

// MockCatalogService
type MockCatalogService struct{}

func (m *MockCatalogService) DeductStock(ctx context.Context, productID string, quantity int) error {
	return nil
}

// MockEventPublisher
type MockEventPublisher struct {
	publishedEvents []string
}

func (m *MockEventPublisher) Publish(ctx context.Context, channel string, event interface{}) error {
	m.publishedEvents = append(m.publishedEvents, channel)
	return nil
}

func TestCheckoutUseCase_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockCartRepository()
	catalogSvc := &MockCatalogService{}
	publisher := &MockEventPublisher{}

	// Monta carrinho no mock
	cart := entity.NewCart("usr-123")
	_ = cart.AddItem("prod-1", "Teclado Mecânico", 29990, 1) // R$ 299,90
	_ = cart.AddItem("prod-2", "Mouse Gamer", 15000, 2)      // 2x R$ 150,00 = 300,00
	_ = cartRepo.Save(context.Background(), cart, 30*time.Minute)

	uc := order.NewCheckoutUseCase(orderRepo, cartRepo, catalogSvc, publisher)

	out, err := uc.Execute(context.Background(), order.CheckoutInput{
		OrderID:        "ord-001",
		UserID:         "usr-123",
		IdempotencyKey: "idem-key-abc-123",
	})

	if err != nil {
		t.Fatalf("esperava sucesso no checkout, obteve: %v", err)
	}

	// Total esperado: 29990 + 30000 = 59990
	if out.TotalAmount != 59990 {
		t.Errorf("total calculado incorreto. Esperava 59990, obteve: %d", out.TotalAmount)
	}

	if out.Status != entity.StatusPending {
		t.Errorf("status inicial incorreto: %s", out.Status)
	}

	// Carrinho deve ter sido limpo
	savedCart, _ := cartRepo.Get(context.Background(), "usr-123")
	if savedCart != nil {
		t.Errorf("o carrinho deveria estar vazio após o checkout")
	}

	// Evento order.criado deve ter sido publicado
	if len(publisher.publishedEvents) != 1 || publisher.publishedEvents[0] != "order.criado" {
		t.Errorf("esperava publicação do evento order.criado, obteve: %v", publisher.publishedEvents)
	}
}

func TestCheckoutUseCase_Idempotency(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockCartRepository()
	catalogSvc := &MockCatalogService{}
	publisher := &MockEventPublisher{}

	cart := entity.NewCart("usr-123")
	_ = cart.AddItem("prod-1", "Headset", 19900, 1)
	_ = cartRepo.Save(context.Background(), cart, 30*time.Minute)

	uc := order.NewCheckoutUseCase(orderRepo, cartRepo, catalogSvc, publisher)

	// Primeira chamada
	out1, err := uc.Execute(context.Background(), order.CheckoutInput{
		OrderID:        "ord-001",
		UserID:         "usr-123",
		IdempotencyKey: "chave-duplicada",
	})
	if err != nil {
		t.Fatalf("erro na 1a chamada: %v", err)
	}

	// Segunda chamada com a MESMA IdempotencyKey
	out2, err := uc.Execute(context.Background(), order.CheckoutInput{
		OrderID:        "ord-002",
		UserID:         "usr-123",
		IdempotencyKey: "chave-duplicada",
	})
	if err != nil {
		t.Fatalf("erro na 2a chamada: %v", err)
	}

	// Deve retornar o MESMO pedido da 1ª chamada
	if out1.ID != out2.ID {
		t.Errorf("idempotência falhou: esperava o mesmo pedido ID %s, mas obteve %s", out1.ID, out2.ID)
	}
}

func TestCheckoutUseCase_EmptyCart(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockCartRepository()
	catalogSvc := &MockCatalogService{}
	publisher := &MockEventPublisher{}

	uc := order.NewCheckoutUseCase(orderRepo, cartRepo, catalogSvc, publisher)

	_, err := uc.Execute(context.Background(), order.CheckoutInput{
		OrderID: "ord-003",
		UserID:  "usr-sem-carrinho",
	})

	if err != order.ErrCartIsEmpty {
		t.Errorf("esperava ErrCartIsEmpty, obteve: %v", err)
	}
}
