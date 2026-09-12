package entity

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidProductID  = errors.New("id de produto inválido")
	ErrEmptyProductName  = errors.New("nome do produto não pode ser vazio")
	ErrInvalidPrice      = errors.New("o preço do produto deve ser maior que zero")
	ErrNegativeStock     = errors.New("o estoque não pode ser negativo")
	ErrInsufficientStock = errors.New("estoque insuficiente para a quantidade solicitada")
	ErrInvalidQuantity   = errors.New("quantidade deve ser maior que zero")
)

type Product struct {
	ID          string
	Name        string
	Slug        string
	Description string
	Price       int64  // Preço em centavos (ex: R$ 99,90 = 9990)
	Stock       int    // Quantidade física disponível
	SKU         string // Código identificador único do item
	CategoryID  string
	ImageURL    string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProduct(
	id, name, description string,
	price int64,
	stock int,
	sku, categoryID, imageURL string,
) (*Product, error) {
	name = strings.TrimSpace(name)
	sku = strings.TrimSpace(strings.ToUpper(sku))

	if id == "" {
		return nil, ErrInvalidProductID
	}
	if name == "" {
		return nil, ErrEmptyProductName
	}
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if stock < 0 {
		return nil, ErrNegativeStock
	}

	slug := generateSlug(name)
	now := time.Now().UTC()

	return &Product{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(description),
		Price:       price,
		Stock:       stock,
		SKU:         sku,
		CategoryID:  categoryID,
		ImageURL:    strings.TrimSpace(imageURL),
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// DeductStock reduz o estoque com validação de limite mínimo
func (p *Product) DeductStock(qty int) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	if p.Stock < qty {
		return ErrInsufficientStock
	}
	p.Stock -= qty
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// AddStock incrementa o estoque
func (p *Product) AddStock(qty int) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	p.Stock += qty
	p.UpdatedAt = time.Now().UTC()
	return nil
}
