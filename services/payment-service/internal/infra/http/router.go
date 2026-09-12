package http

import (
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *handler.PaymentHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Signature-SHA256"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.HealthCheck)

	r.Route("/payments", func(r chi.Router) {
		r.Post("/webhook", h.HandleWebhook)
	})

	r.Route("/financial", func(r chi.Router) {
		r.Get("/summary", h.GetSummary)
		r.Get("/transactions", h.ListTransactions)
	})

	return r
}
