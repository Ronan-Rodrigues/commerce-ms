package http

import (
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/infra/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *handler.NotifyHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.HealthCheck)
	r.Get("/events", h.StreamEvents)
	r.Post("/webhook/confirmar", h.HandleConfirmation)

	return r
}
