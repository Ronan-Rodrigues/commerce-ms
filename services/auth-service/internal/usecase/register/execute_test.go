package register_test

import (
	"context"
	"testing"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/register"
)

// MockUserRepository simula o repositório em memória
type MockUserRepository struct {
	users map[string]*entity.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{users: make(map[string]*entity.User)}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

// MockPasswordHasher simula o hasher
type MockPasswordHasher struct{}

func (m *MockPasswordHasher) HashPassword(p string) (string, error) {
	return "hashed_" + p, nil
}

func (m *MockPasswordHasher) ComparePassword(p, hash string) bool {
	return "hashed_"+p == hash
}

func TestRegisterUseCase_Success(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := &MockPasswordHasher{}
	uc := register.NewUseCase(repo, hasher)

	input := register.InputDTO{
		ID:       "usr-1",
		Name:     "Ronan Rodrigues",
		Email:    "ronan@example.com",
		Password: "senhaSegura123",
		Role:     "customer",
	}

	out, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve erro: %v", err)
	}

	if out.ID != "usr-1" || out.Email != "ronan@example.com" {
		t.Errorf("dados retornados inconsistentes: %+v", out)
	}
}

func TestRegisterUseCase_PasswordTooShort(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := &MockPasswordHasher{}
	uc := register.NewUseCase(repo, hasher)

	input := register.InputDTO{
		ID:       "usr-2",
		Name:     "Ronan Rodrigues",
		Email:    "ronan@example.com",
		Password: "123", // menos de 8 chars
	}

	_, err := uc.Execute(context.Background(), input)
	if err != register.ErrPasswordTooShort {
		t.Errorf("esperava erro ErrPasswordTooShort, obteve: %v", err)
	}
}

func TestRegisterUseCase_UserAlreadyExists(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := &MockPasswordHasher{}
	uc := register.NewUseCase(repo, hasher)

	input := register.InputDTO{
		ID:       "usr-3",
		Name:     "Ronan Rodrigues",
		Email:    "duplicado@example.com",
		Password: "senhaSegura123",
	}

	// Primeiro registro
	_, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("erro no primeiro registro: %v", err)
	}

	// Segundo registro com mesmo email deve falhar
	_, err = uc.Execute(context.Background(), input)
	if err != register.ErrUserAlreadyExists {
		t.Errorf("esperava erro ErrUserAlreadyExists, obteve: %v", err)
	}
}

