package internal

import (
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type GatewayProxies struct {
	Auth    http.Handler
	Catalog http.Handler
	Order   http.Handler
	Payment http.Handler
	Notify  http.Handler
}

func NewRouter(
	proxies GatewayProxies,
	authMW *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
) http.Handler {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(rateLimiter.Middleware)

	// Headers de segurança
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			next.ServeHTTP(w, r)
		})
	})

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Check do Gateway
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok", "service": "api-gateway"}`))
	})

	// ── 1. Rotas Públicas ──────────────────────────────────────────

	// Autenticação (cadastro, login, refresh)
	r.Mount("/auth", proxies.Auth)

	// Catálogo (leitura pública de produtos e categorias)
	r.Get("/products*", proxies.Catalog.ServeHTTP)
	r.Get("/categories*", proxies.Catalog.ServeHTTP)

	// Webhooks de pagamento (validados por assinatura HMAC no próprio serviço)
	r.Post("/payments/webhook", proxies.Payment.ServeHTTP)

	// ── 2. Rotas Protegidas (Requer JWT) ───────────────────────────
	r.Group(func(protected chi.Router) {
		protected.Use(authMW.RequireAuth)

		// Operações administrativas de catálogo
		protected.Post("/products*", proxies.Catalog.ServeHTTP)
		protected.Put("/products*", proxies.Catalog.ServeHTTP)
		protected.Delete("/products*", proxies.Catalog.ServeHTTP)

		// Carrinho e Pedidos
		protected.Mount("/cart", proxies.Order)
		protected.Mount("/orders", proxies.Order)

		// Dashboard Financeiro
		protected.Mount("/financial", proxies.Payment)

		// Notificações SSE em tempo real
		protected.Get("/events", proxies.Notify.ServeHTTP)
	})

	return r
}
