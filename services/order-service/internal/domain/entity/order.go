package entity

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pendente"
	StatusPaid      OrderStatus = "pago"
	StatusShipped   OrderStatus = "enviado"
	StatusDelivered OrderStatus = "entregue"
	StatusCanceled  OrderStatus = "cancelado"
)

var (
	ErrEmptyCart          = errors.New("não é possível criar um pedido com carrinho vazio")
	ErrInvalidStatusOrder = errors.New("transição de status de pedido inválida")
)

type OrderItem struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"` // em centavos
	Quantity  int    `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

type Order struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	TotalAmount    int64        `json:"total_amount"`
	Status         OrderStatus  `json:"status"`
	IdempotencyKey string       `json:"idempotency_key"`
	Items          []*OrderItem `json:"items"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func NewOrder(id, userID, idempotencyKey string, cart *Cart) (*Order, error) {
	if cart == nil || len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	now := time.Now().UTC()
	var items []*OrderItem

	for _, cItem := range cart.Items {
		items = append(items, &OrderItem{
			ProductID: cItem.ProductID,
			Name:      cItem.Name,
			Price:     cItem.Price,
			Quantity:  cItem.Quantity,
			Subtotal:  cItem.Subtotal,
		})
	}

	return &Order{
		ID:             id,
		UserID:         userID,
		TotalAmount:    cart.TotalAmount,
		Status:         StatusPending,
		IdempotencyKey: idempotencyKey,
		Items:          items,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// TransitionTo valida transição de status usando máquina de estados
func (o *Order) TransitionTo(newStatus OrderStatus) error {
	switch o.Status {
	case StatusPending:
		if newStatus == StatusPaid || newStatus == StatusCanceled {
			o.Status = newStatus
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	case StatusPaid:
		if newStatus == StatusShipped || newStatus == StatusCanceled {
			o.Status = newStatus
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	case StatusShipped:
		if newStatus == StatusDelivered {
			o.Status = newStatus
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return ErrInvalidStatusOrder
}
