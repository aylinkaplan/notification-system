package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aylinkaplan/notification-system/internal/api"
	"github.com/aylinkaplan/notification-system/internal/api/handler"
	"github.com/aylinkaplan/notification-system/internal/config"
	"github.com/aylinkaplan/notification-system/internal/provider"
	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/aylinkaplan/notification-system/internal/ratelimit"
	"github.com/aylinkaplan/notification-system/internal/repository"
	"github.com/aylinkaplan/notification-system/internal/service"
	"github.com/aylinkaplan/notification-system/internal/storage"
	"github.com/aylinkaplan/notification-system/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	if err := storage.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	q, err := queue.NewRabbitMQQueue(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer q.Close()

	notificationRepo := repository.NewNotificationRepository(db)
	enqueuer := queue.NewEnqueuerAdapter(q)
	notificationSvc := service.NewNotificationService(notificationRepo, enqueuer)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)

	prov := provider.NewWebhookProvider(cfg.WebhookBaseURL)
	limiter := ratelimit.NewMemoryLimiter(cfg.RateLimitPerSec)
	w := worker.NewWorker(q, notificationRepo, prov, limiter, cfg.WebhookUUID)
	ctx := context.Background()
	w.Run(ctx)

	metricsHandler := handler.NewMetricsHandler(&handler.MetricsDB{DB: db}, q)
	router := api.NewRouter(cfg, db, q, notificationHandler, metricsHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	log.Info().Msg("server stopped")
}
