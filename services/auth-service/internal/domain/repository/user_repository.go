package repository

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
)

var (
	ErrUserNotFound      = errors.New("usuário não encontrado")
	ErrUserAlreadyExists = errors.New("usuário com este email já cadastrado")
)

// UserRepository define o contrato que qualquer mecanismo de persistência deve implementar
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
}

