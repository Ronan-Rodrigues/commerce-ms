package entity

import (
	"errors"
	"strings"
	"time"
)

// Erros de domínio relacionados ao usuário
var (
	ErrInvalidUserID    = errors.New("id de usuário inválido")
	ErrInvalidEmail     = errors.New("formato de email inválido")
	ErrPasswordTooShort = errors.New("a senha deve ter no mínimo 8 caracteres")
	ErrEmptyName        = errors.New("o nome não pode ser vazio")
)

// User representa a entidade central de usuário no domínio
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string // "customer" ou "admin"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser cria e valida uma nova entidade User
func NewUser(id, name, email, passwordHash, role string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	if id == "" {
		return nil, ErrInvalidUserID
	}
	if name == "" {
		return nil, ErrEmptyName
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, ErrInvalidEmail
	}
	if role == "" {
		role = "customer"
	}

	now := time.Now().UTC()

	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

