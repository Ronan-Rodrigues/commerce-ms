package cart

import (
	"context"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/repository"
)

type AddItemInput struct {
	UserID    string
	ProductID string
	Name      string
	Price     int64
	Quantity  int
}

type CartUseCase struct {
	cartRepo repository.CartRepository
	ttl      time.Duration
}

func NewCartUseCase(cartRepo repository.CartRepository, ttl time.Duration) *CartUseCase {
	if ttl <= 0 {
		ttl = 30 * time.Minute // 30 minutos de retenção no carrinho
	}
	return &CartUseCase{
		cartRepo: cartRepo,
		ttl:      ttl,
	}
}

func (uc *CartUseCase) AddItem(ctx context.Context, input AddItemInput) (*entity.Cart, error) {
	currentCart, err := uc.cartRepo.Get(ctx, input.UserID)
	if err != nil || currentCart == nil {
		currentCart = entity.NewCart(input.UserID)
	}

	if err := currentCart.AddItem(input.ProductID, input.Name, input.Price, input.Quantity); err != nil {
		return nil, err
	}

	if err := uc.cartRepo.Save(ctx, currentCart, uc.ttl); err != nil {
		return nil, err
	}

	return currentCart, nil
}

func (uc *CartUseCase) GetCart(ctx context.Context, userID string) (*entity.Cart, error) {
	currentCart, err := uc.cartRepo.Get(ctx, userID)
	if err != nil || currentCart == nil {
		return entity.NewCart(userID), nil
	}
	return currentCart, nil
}

func (uc *CartUseCase) RemoveItem(ctx context.Context, userID, productID string) (*entity.Cart, error) {
	currentCart, err := uc.cartRepo.Get(ctx, userID)
	if err != nil || currentCart == nil {
		return nil, entity.ErrCartItemNotFound
	}

	if err := currentCart.RemoveItem(productID); err != nil {
		return nil, err
	}

	if err := uc.cartRepo.Save(ctx, currentCart, uc.ttl); err != nil {
		return nil, err
	}

	return currentCart, nil
}
