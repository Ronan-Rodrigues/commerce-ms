package refresh_test

import (
	"context"
	"testing"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/refresh"
)

type MockUserRepository struct {
	users map[string]*entity.User
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

type MockRefreshStore struct {
	tokens map[string]string // refreshToken -> userID
}

func NewMockRefreshStore() *MockRefreshStore {
	return &MockRefreshStore{tokens: make(map[string]string)}
}

func (m *MockRefreshStore) ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if uid, ok := m.tokens[refreshToken]; ok {
		return uid, nil
	}
	return "", nil
}

func (m *MockRefreshStore) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	delete(m.tokens, refreshToken)
	return nil
}

func (m *MockRefreshStore) StoreRefreshToken(ctx context.Context, refreshToken, userID string, ttl time.Duration) error {
	m.tokens[refreshToken] = userID
	return nil
}

type MockTokenService struct{}

func (m *MockTokenService) GenerateAccessToken(payload service.TokenPayload, ttl time.Duration) (string, error) {
	return "new.access.token", nil
}

func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	return "new-refresh-token-uuid", nil
}

func TestRefreshUseCase_Success(t *testing.T) {
	repo := &MockUserRepository{users: make(map[string]*entity.User)}
	store := NewMockRefreshStore()
	tokenSvc := &MockTokenService{}

	user, _ := entity.NewUser("usr-1", "Ronan", "ronan@example.com", "hash", "customer")
	_ = repo.Create(context.Background(), user)
	_ = store.StoreRefreshToken(context.Background(), "valid-token-123", user.ID, 24*time.Hour)

	uc := refresh.NewUseCase(repo, store, tokenSvc, refresh.TokenConfig{
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	})

	out, err := uc.Execute(context.Background(), refresh.InputDTO{
		RefreshToken: "valid-token-123",
	})

	if err != nil {
		t.Fatalf("esperava sucesso na rotação de token, obteve: %v", err)
	}

	if out.AccessToken != "new.access.token" {
		t.Errorf("access token incorreto: %s", out.AccessToken)
	}
	if out.RefreshToken != "new-refresh-token-uuid" {
		t.Errorf("refresh token incorreto: %s", out.RefreshToken)
	}

	// Verifica se o token anterior foi deletado do store (Single-Use Rotation)
	oldValid, _ := store.ValidateRefreshToken(context.Background(), "valid-token-123")
	if oldValid != "" {
		t.Errorf("o refresh token antigo deveria ter sido invalidado na rotação")
	}

	// Verifica se o novo token foi gravado
	newValid, _ := store.ValidateRefreshToken(context.Background(), "new-refresh-token-uuid")
	if newValid != user.ID {
		t.Errorf("novo refresh token deveria estar associado ao usuário")
	}
}

func TestRefreshUseCase_InvalidToken(t *testing.T) {
	repo := &MockUserRepository{users: make(map[string]*entity.User)}
	store := NewMockRefreshStore()
	tokenSvc := &MockTokenService{}

	uc := refresh.NewUseCase(repo, store, tokenSvc, refresh.TokenConfig{
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	})

	_, err := uc.Execute(context.Background(), refresh.InputDTO{
		RefreshToken: "token-inexistente",
	})

	if err != refresh.ErrInvalidRefreshToken {
		t.Errorf("esperava ErrInvalidRefreshToken, obteve: %v", err)
	}
}
