package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aylinkaplan/notification-system/internal/config"
	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/jmoiron/sqlx"
)

func HealthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	}
}

func ReadyHandler(cfg *config.Config, db *sqlx.DB, q queue.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		checks := map[string]string{
			"database":  "ok",
			"rabbitmq": "ok",
		}

		if err := db.PingContext(ctx); err != nil {
			checks["database"] = err.Error()
		}
		if _, err := q.AllDepths(ctx); err != nil {
			checks["rabbitmq"] = err.Error()
		}

		allOk := checks["database"] == "ok" && checks["rabbitmq"] == "ok"

		w.Header().Set("Content-Type", "application/json")
		if !allOk {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": map[bool]string{true: "ready", false: "not_ready"}[allOk],
			"checks": checks,
		})
	}
}
