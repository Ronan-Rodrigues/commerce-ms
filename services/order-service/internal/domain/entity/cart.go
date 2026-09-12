package entity

import (
	"errors"
	"time"
)

var (
	ErrInvalidQuantity  = errors.New("quantidade do item deve ser maior que zero")
	ErrCartItemNotFound = errors.New("item não encontrado no carrinho")
)

type CartItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"` // em centavos
	Quantity  int    `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

type Cart struct {
	UserID      string      `json:"user_id"`
	Items       []*CartItem `json:"items"`
	TotalAmount int64       `json:"total_amount"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func NewCart(userID string) *Cart {
	return &Cart{
		UserID:      userID,
		Items:       []*CartItem{},
		TotalAmount: 0,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (c *Cart) AddItem(productID, name string, price int64, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	for _, item := range c.Items {
		if item.ProductID == productID {
			item.Quantity += quantity
			item.Subtotal = int64(item.Quantity) * item.Price
			c.recalculateTotal()
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	newItem := &CartItem{
		ProductID: productID,
		Name:      name,
		Price:     price,
		Quantity:  quantity,
		Subtotal:  int64(quantity) * price,
	}

	c.Items = append(c.Items, newItem)
	c.recalculateTotal()
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Cart) RemoveItem(productID string) error {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.recalculateTotal()
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return ErrCartItemNotFound
}

func (c *Cart) recalculateTotal() {
	var total int64
	for _, item := range c.Items {
		total += item.Subtotal
	}
	c.TotalAmount = total
}
