package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	publicKey *rsa.PublicKey
	rdb       *redis.Client
}

func NewAuthMiddleware(publicKey *rsa.PublicKey, rdb *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		publicKey: publicKey,
		rdb:       rdb,
	}
}

func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "autenticação obrigatória: token ausente"}`))
			return
		}

		claims, err := a.validateToken(tokenStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "token inválido ou expirado"}`))
			return
		}

		// Verifica se o token consta na blacklist do Redis
		if claims.ID != "" {
			key := fmt.Sprintf("blacklist:%s", claims.ID)
			exists, err := a.rdb.Exists(r.Context(), key).Result()
			if err == nil && exists > 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "token foi revogado"}`))
				return
			}
		}

		// Injeta os dados do usuário autenticado para os microsserviços internos
		r.Header.Set("X-User-Id", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)

		next.ServeHTTP(w, r)
	})
}

func (a *AuthMiddleware) validateToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("algoritmo de token inválido")
		}
		return a.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}

func extractToken(r *http.Request) string {
	// 1. Tenta header Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. Tenta cookie httpOnly
	cookie, err := r.Cookie("jwt_access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

