package http

import (
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *handler.OrderHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Id", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.HealthCheck)

	// Rotas de Carrinho
	r.Route("/cart", func(r chi.Router) {
		r.Get("/", h.GetCart)
		r.Post("/items", h.AddCartItem)
		r.Delete("/items/{productId}", h.RemoveCartItem)
	})

	// Rotas de Pedidos
	r.Route("/orders", func(r chi.Router) {
		r.Post("/checkout", h.Checkout)
		r.Get("/", h.ListOrders)
		r.Get("/{id}", h.GetOrderByID)
	})

	return r
}
