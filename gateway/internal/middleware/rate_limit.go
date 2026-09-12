package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb            *redis.Client
	limitPerMinute int64
}

func NewRateLimiter(rdb *redis.Client, limitPerMinute int64) *RateLimiter {
	if limitPerMinute <= 0 {
		limitPerMinute = 100 // padrão: 100 req/min por IP
	}
	return &RateLimiter{
		rdb:            rdb,
		limitPerMinute: limitPerMinute,
	}
}

// Middleware aplica rate limiting por IP usando Redis Sliding Window
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.RemoteAddr
		}

		key := fmt.Sprintf("rate:%s", ip)
		now := time.Now().UnixMilli()
		windowStart := now - 60000 // 1 minuto atrás

		ctx := r.Context()

		// Pipeline atômica no Redis para performance máxima
		pipe := rl.rdb.Pipeline()
		// Remove registros mais velhos que 1 minuto
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
		// Adiciona a requisição atual com score igual ao timestamp atual
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: strconv.FormatInt(now, 10)})
		// Conta requisições na janela atual
		countCmd := pipe.ZCard(ctx, key)
		// Renova expiração da chave para não poluir o Redis
		pipe.Expire(ctx, key, 70*time.Second)

		_, err := pipe.Exec(ctx)
		if err != nil {
			// Se o Redis falhar, fail-open (deixa passar com log para não travar a aplicação)
			next.ServeHTTP(w, r)
			return
		}

		currentCount := countCmd.Val()
		if currentCount > rl.limitPerMinute {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "taxa de requisições excedida. Tente novamente em 1 minuto."}`))
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rl.limitPerMinute, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(rl.limitPerMinute-currentCount, 10))

		next.ServeHTTP(w, r)
	})
}
