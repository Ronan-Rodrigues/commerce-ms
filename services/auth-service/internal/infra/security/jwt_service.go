package security

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken   = errors.New("token inválido ou expirado")
	ErrInvalidKeyType = errors.New("tipo de chave de assinatura inválido")
)

// CustomClaims define o formato dos dados contidos no token JWT
type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService implementa a interface service.TokenService usando o algoritmo RS256
type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

// NewJWTService cria uma nova instância do JWTService
func NewJWTService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer string) *JWTService {
	if issuer == "" {
		issuer = "commerce-ms"
	}
	return &JWTService{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
	}
}

// GenerateAccessToken assina um token com a chave privada RSA
func (s *JWTService) GenerateAccessToken(payload service.TokenPayload, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	tokenID := uuid.NewString()

	claims := CustomClaims{
		UserID: payload.UserID,
		Email:  payload.Email,
		Role:   payload.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   payload.UserID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GenerateRefreshToken gera uma string aleatória de alta entropia (32 bytes = 64 caracteres hex)
func (s *JWTService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateAccessToken verifica a assinatura e validade do token usando a chave pública RSA
func (s *JWTService) ValidateAccessToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidKeyType
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
