package http

import (
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/infra/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(productHandler *handler.ProductHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", productHandler.HealthCheck)

	r.Route("/products", func(r chi.Router) {
		r.Get("/", productHandler.List)
		r.Post("/", productHandler.Create)
		r.Get("/{id}", productHandler.GetByID)
		r.Post("/{id}/deduct-stock", productHandler.DeductStock)
	})

	return r
}
