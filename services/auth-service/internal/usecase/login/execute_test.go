package login_test

import (
	"context"
	"testing"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/login"
)

// MockUserRepository
type MockUserRepository struct {
	users map[string]*entity.User
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

// MockPasswordHasher
type MockPasswordHasher struct{}

func (m *MockPasswordHasher) HashPassword(p string) (string, error) {
	return "hashed_" + p, nil
}

func (m *MockPasswordHasher) ComparePassword(p, hash string) bool {
	return "hashed_"+p == hash
}

// MockTokenService
type MockTokenService struct{}

func (m *MockTokenService) GenerateAccessToken(payload service.TokenPayload, ttl time.Duration) (string, error) {
	return "mock.jwt.token", nil
}

func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	return "mock-refresh-token-uuid", nil
}

func setupTest() (*login.UseCase, *MockUserRepository) {
	repo := &MockUserRepository{users: make(map[string]*entity.User)}
	hasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	tokenConfig := login.TokenConfig{
		AccessTTL: 15 * time.Minute,
	}

	uc := login.NewUseCase(repo, hasher, tokenService, tokenConfig)
	return uc, repo
}

func TestLoginUseCase_Success(t *testing.T) {
	uc, repo := setupTest()

	// Cadastra usuário prévio
	user, _ := entity.NewUser("usr-1", "Ronan", "ronan@example.com", "hashed_correta123", "customer")
	_ = repo.Create(context.Background(), user)

	input := login.InputDTO{
		Email:    "ronan@example.com",
		Password: "correta123",
	}

	out, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("esperava sucesso no login, obteve: %v", err)
	}

	if out.AccessToken != "mock.jwt.token" {
		t.Errorf("token de acesso incorreto: %s", out.AccessToken)
	}
	if out.RefreshToken != "mock-refresh-token-uuid" {
		t.Errorf("refresh token incorreto: %s", out.RefreshToken)
	}
	if out.User.ID != "usr-1" || out.User.Email != "ronan@example.com" {
		t.Errorf("dados de usuário inconsistentes: %+v", out.User)
	}
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	uc, repo := setupTest()

	user, _ := entity.NewUser("usr-1", "Ronan", "ronan@example.com", "hashed_correta123", "customer")
	_ = repo.Create(context.Background(), user)

	input := login.InputDTO{
		Email:    "ronan@example.com",
		Password: "senhaErrada",
	}

	_, err := uc.Execute(context.Background(), input)
	if err != login.ErrInvalidCredentials {
		t.Errorf("esperava ErrInvalidCredentials, obteve: %v", err)
	}
}

func TestLoginUseCase_UserNotFound(t *testing.T) {
	uc, _ := setupTest()

	input := login.InputDTO{
		Email:    "naoexiste@example.com",
		Password: "qualquerSenha",
	}

	_, err := uc.Execute(context.Background(), input)
	if err != login.ErrInvalidCredentials {
		t.Errorf("esperava ErrInvalidCredentials, obteve: %v", err)
	}
}

func TestLoginUseCase_EmptyFields(t *testing.T) {
	uc, _ := setupTest()

	input := login.InputDTO{
		Email:    "",
		Password: "",
	}

	_, err := uc.Execute(context.Background(), input)
	if err != login.ErrInvalidCredentials {
		t.Errorf("esperava ErrInvalidCredentials para campos vazios, obteve: %v", err)
	}
}

