package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/repository"
)

var (
	ErrCartIsEmpty = errors.New("não é possível finalizar compra com carrinho vazio")
)

type CheckoutInput struct {
	OrderID        string
	UserID         string
	IdempotencyKey string
}

type OrderCreatedEvent struct {
	OrderID     string              `json:"order_id"`
	UserID      string              `json:"user_id"`
	TotalAmount int64               `json:"total_amount"`
	Status      string              `json:"status"`
	Items       []*entity.OrderItem `json:"items"`
}

type CheckoutUseCase struct {
	orderRepo  repository.OrderRepository
	cartRepo   repository.CartRepository
	catalogSvc repository.CatalogService
	publisher  repository.EventPublisher
}

func NewCheckoutUseCase(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	catalogSvc repository.CatalogService,
	publisher repository.EventPublisher,
) *CheckoutUseCase {
	return &CheckoutUseCase{
		orderRepo:  orderRepo,
		cartRepo:   cartRepo,
		catalogSvc: catalogSvc,
		publisher:  publisher,
	}
}

func (uc *CheckoutUseCase) Execute(ctx context.Context, input CheckoutInput) (*entity.Order, error) {
	// 1. Verificação de Idempotência: Se o pedido com esta chave já foi criado, retorne-o
	if input.IdempotencyKey != "" {
		existingOrder, err := uc.orderRepo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil && existingOrder != nil {
			return existingOrder, nil
		}
	}

	// 2. Busca o carrinho atual do usuário
	cart, err := uc.cartRepo.Get(ctx, input.UserID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		return nil, ErrCartIsEmpty
	}

	// 3. Dedução atômica de estoque para cada item no catálogo
	for _, item := range cart.Items {
		if err := uc.catalogSvc.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
			return nil, fmt.Errorf("falha ao reservar estoque do produto %s: %w", item.Name, err)
		}
	}

	// 4. Cria o Pedido de domínio
	order, err := entity.NewOrder(input.OrderID, input.UserID, input.IdempotencyKey, cart)
	if err != nil {
		return nil, err
	}

	// 5. Salva o pedido no banco de dados (Postgres)
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 6. Limpa o carrinho no Redis
	_ = uc.cartRepo.Clear(ctx, input.UserID)

	// 7. Publica o evento order.criado no Redis Pub/Sub
	event := OrderCreatedEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		Items:       order.Items,
	}
	_ = uc.publisher.Publish(ctx, "order.criado", event)

	return order, nil
}
