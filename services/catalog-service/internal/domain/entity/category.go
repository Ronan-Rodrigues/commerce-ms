package entity

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidCategoryID = errors.New("id de categoria inválido")
	ErrEmptyCategoryName = errors.New("o nome da categoria não pode ser vazio")
)

type Category struct {
	ID          string
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCategory(id, name, description string) (*Category, error) {
	name = strings.TrimSpace(name)
	if id == "" {
		return nil, ErrInvalidCategoryID
	}
	if name == "" {
		return nil, ErrEmptyCategoryName
	}

	slug := generateSlug(name)
	now := time.Now().UTC()

	return &Category{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// generateSlug converte "Eletrônicos & Informática" em "eletronicos-informatica"
func generateSlug(s string) string {
	s = strings.ToLower(s)
	// Remove acentos comuns em pt-br
	replacer := strings.NewReplacer(
		"ã", "a", "á", "a", "à", "a", "â", "a",
		"é", "e", "ê", "e",
		"í", "i",
		"ó", "o", "õ", "o", "ô", "o",
		"ú", "u", "ü", "u",
		"ç", "c",
	)
	s = replacer.Replace(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
