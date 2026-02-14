package api

import (
	"net/http"

	"github.com/aylinkaplan/notification-system/internal/api/handler"
	"github.com/aylinkaplan/notification-system/internal/api/middleware"
	"github.com/aylinkaplan/notification-system/internal/config"
	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func NewRouter(cfg *config.Config, db *sqlx.DB, q queue.Queue, notificationHandler *handler.NotificationHandler, metricsHandler *handler.MetricsHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CorrelationID)

	r.Get("/health", HealthHandler(cfg))
	r.Get("/ready", ReadyHandler(cfg, db, q))
	r.Get("/metrics", metricsHandler.ServeHTTP)
	r.Get("/openapi.yaml", OpenAPIHandler)
	r.Get("/docs", DocsHandler)
	r.Get("/docs/", DocsHandler)

	r.Route("/notifications", func(r chi.Router) {
		r.Post("/", notificationHandler.Create)
		r.Post("/batch", notificationHandler.CreateBatch)
		r.Get("/", notificationHandler.List)
		r.Get("/batch/{batchId}", notificationHandler.GetByBatchID)
		r.Get("/{id}", notificationHandler.GetByID)
		r.Delete("/{id}", notificationHandler.Cancel)
	})

	return r
}
