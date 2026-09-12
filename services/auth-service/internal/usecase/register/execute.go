package register

import (
	"context"
	"errors"
	"strings"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
)

var (
	ErrUserAlreadyExists = errors.New("já existe um usuário cadastrado com este email")
	ErrPasswordTooShort  = errors.New("a senha deve conter pelo menos 8 caracteres")
)

// InputDTO contém os dados de entrada para o registro de usuário
type InputDTO struct {
	ID       string
	Name     string
	Email    string
	Password string
	Role     string
}

// OutputDTO contém os dados de saída retornados após o registro
type OutputDTO struct {
	ID    string
	Name  string
	Email string
	Role  string
}

// UseCase orquestra o fluxo de criação de conta
type UseCase struct {
	userRepo repository.UserRepository
	hasher   service.PasswordHasher
}

// NewUseCase instancia o caso de uso com injeção de dependências
func NewUseCase(userRepo repository.UserRepository, hasher service.PasswordHasher) *UseCase {
	return &UseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// Execute executa o caso de uso
func (uc *UseCase) Execute(ctx context.Context, input InputDTO) (*OutputDTO, error) {
	if len(strings.TrimSpace(input.Password)) < 8 {
		return nil, ErrPasswordTooShort
	}

	cleanEmail := strings.TrimSpace(strings.ToLower(input.Email))

	existingUser, err := uc.userRepo.FindByEmail(ctx, cleanEmail)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := uc.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	newUser, err := entity.NewUser(input.ID, input.Name, cleanEmail, passwordHash, input.Role)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return &OutputDTO{
		ID:    newUser.ID,
		Name:  newUser.Name,
		Email: newUser.Email,
		Role:  newUser.Role,
	}, nil
}

