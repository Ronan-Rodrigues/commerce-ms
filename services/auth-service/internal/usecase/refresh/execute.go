package refresh

import (
	"context"
	"errors"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
)

var (
	ErrInvalidRefreshToken = errors.New("refresh token inválido ou expirado")
)

// RefreshStore define as operações necessárias do Redis para este usecase
type RefreshStore interface {
	ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	StoreRefreshToken(ctx context.Context, refreshToken, userID string, ttl time.Duration) error
}

type InputDTO struct {
	RefreshToken string
}

type OutputDTO struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type TokenConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type UseCase struct {
	userRepo     repository.UserRepository
	refreshStore RefreshStore
	tokenService service.TokenService
	tokenConfig  TokenConfig
}

func NewUseCase(
	userRepo repository.UserRepository,
	refreshStore RefreshStore,
	tokenService service.TokenService,
	tokenConfig TokenConfig,
) *UseCase {
	return &UseCase{
		userRepo:     userRepo,
		refreshStore: refreshStore,
		tokenService: tokenService,
		tokenConfig:  tokenConfig,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input InputDTO) (*OutputDTO, error) {
	if input.RefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	// 1. Valida refresh token no store
	userID, err := uc.refreshStore.ValidateRefreshToken(ctx, input.RefreshToken)
	if err != nil || userID == "" {
		return nil, ErrInvalidRefreshToken
	}

	// 2. Rotação: Deleta o refresh token antigo imediatamente (single-use)
	_ = uc.refreshStore.DeleteRefreshToken(ctx, input.RefreshToken)

	// 3. Obtém os dados mais atualizados do usuário
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// 4. Emite novo access token
	payload := service.TokenPayload{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}
	newAccessToken, err := uc.tokenService.GenerateAccessToken(payload, uc.tokenConfig.AccessTTL)
	if err != nil {
		return nil, err
	}

	// 5. Emite novo refresh token
	newRefreshToken, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 6. Armazena o novo refresh token
	err = uc.refreshStore.StoreRefreshToken(ctx, newRefreshToken, user.ID, uc.tokenConfig.RefreshTTL)
	if err != nil {
		return nil, err
	}

	return &OutputDTO{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(uc.tokenConfig.AccessTTL.Seconds()),
	}, nil
}
