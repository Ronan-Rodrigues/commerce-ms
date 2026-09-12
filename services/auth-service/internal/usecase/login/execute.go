package login

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrAccountInactive    = errors.New("conta inativa ou bloqueada")
)

// InputDTO dados necessários para autenticação
type InputDTO struct {
	Email    string
	Password string
}

// OutputDTO dados retornados após login bem-sucedido
type OutputDTO struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // segundos até expirar o access token
	User         UserOutputDTO
}

type UserOutputDTO struct {
	ID    string
	Name  string
	Email string
	Role  string
}

// Configs de expiração dos tokens
type TokenConfig struct {
	AccessTTL time.Duration
}

// UseCase coordena o fluxo de login
type UseCase struct {
	userRepo     repository.UserRepository
	hasher       service.PasswordHasher
	tokenService service.TokenService
	tokenConfig  TokenConfig
}

// NewUseCase instancia o UseCase com suas dependências
func NewUseCase(
	userRepo repository.UserRepository,
	hasher service.PasswordHasher,
	tokenService service.TokenService,
	tokenConfig TokenConfig,
) *UseCase {
	return &UseCase{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenService: tokenService,
		tokenConfig:  tokenConfig,
	}
}

// Execute executa o login do usuário
func (uc *UseCase) Execute(ctx context.Context, input InputDTO) (*OutputDTO, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(input.Email))
	if cleanEmail == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := uc.userRepo.FindByEmail(ctx, cleanEmail)
	if err != nil {
		// Não revela se o email existe ou não
		return nil, ErrInvalidCredentials
	}

	if !uc.hasher.ComparePassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	payload := service.TokenPayload{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(payload, uc.tokenConfig.AccessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &OutputDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(uc.tokenConfig.AccessTTL.Seconds()),
		User: UserOutputDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

