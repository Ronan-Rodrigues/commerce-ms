package security_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/service"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/security"
)

func generateTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("erro ao gerar chave RSA de teste: %v", err)
	}
	return privKey, &privKey.PublicKey
}

func TestJWTService_GenerateAndValidate(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)
	jwtSvc := security.NewJWTService(privKey, pubKey, "test-issuer")

	payload := service.TokenPayload{
		UserID: "usr-123",
		Email:  "ronan@example.com",
		Role:   "admin",
	}

	// 1. Gera token com 10 minutos de vida
	tokenStr, err := jwtSvc.GenerateAccessToken(payload, 10*time.Minute)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("token gerado está vazio")
	}

	// 2. Valida o token com a chave pública
	claims, err := jwtSvc.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("falha ao validar token válido: %v", err)
	}

	if claims.UserID != "usr-123" {
		t.Errorf("esperava UserID usr-123, obteve: %s", claims.UserID)
	}
	if claims.Email != "ronan@example.com" {
		t.Errorf("esperava Email ronan@example.com, obteve: %s", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("esperava Role admin, obteve: %s", claims.Role)
	}
	if claims.ID == "" {
		t.Errorf("esperava que jti (token ID) estivesse preenchido")
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)
	jwtSvc := security.NewJWTService(privKey, pubKey, "test-issuer")

	payload := service.TokenPayload{
		UserID: "usr-123",
		Email:  "ronan@example.com",
		Role:   "customer",
	}

	// Token com TTL negativo (já nasce expirado)
	tokenStr, err := jwtSvc.GenerateAccessToken(payload, -1*time.Minute)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	_, err = jwtSvc.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("esperava erro de token expirado, mas validação passou")
	}
}

func TestJWTService_TamperedToken(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)
	jwtSvc := security.NewJWTService(privKey, pubKey, "test-issuer")

	payload := service.TokenPayload{
		UserID: "usr-123",
		Email:  "ronan@example.com",
		Role:   "customer",
	}

	tokenStr, _ := jwtSvc.GenerateAccessToken(payload, 15*time.Minute)

	// Adulterando a assinatura no final da string do token
	tamperedToken := tokenStr[:len(tokenStr)-4] + "ABCD"

	_, err := jwtSvc.ValidateAccessToken(tamperedToken)
	if err == nil {
		t.Fatal("esperava erro para token adulterado, mas validação passou")
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)
	jwtSvc := security.NewJWTService(privKey, pubKey, "test-issuer")

	t1, err := jwtSvc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("falha ao gerar refresh token 1: %v", err)
	}

	t2, err := jwtSvc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("falha ao gerar refresh token 2: %v", err)
	}

	if len(t1) != 64 || len(t2) != 64 {
		t.Errorf("refresh token deve ter 64 caracteres hexadecimais")
	}

	if t1 == t2 {
		t.Errorf("dois refresh tokens não podem ser idênticos")
	}
}
