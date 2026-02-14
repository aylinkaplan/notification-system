package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/jmoiron/sqlx"
)

// NotificationCountReader returns notification counts by status (for metrics).
type NotificationCountReader interface {
	CountByStatus(ctx context.Context, status string) (int64, error)
}

type MetricsHandler struct {
	db    NotificationCountReader
	queue queue.Queue
}

// NewMetricsHandler builds a metrics handler. db can be *sqlx.DB or a test double.
func NewMetricsHandler(db NotificationCountReader, q queue.Queue) *MetricsHandler {
	return &MetricsHandler{db: db, queue: q}
}

// MetricsDB adapts *sqlx.DB to NotificationCountReader.
type MetricsDB struct{ DB *sqlx.DB }

func (m *MetricsDB) CountByStatus(ctx context.Context, status string) (int64, error) {
	var n int64
	err := m.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE status = $1`, status).Scan(&n)
	return n, err
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	depths, err := h.queue.AllDepths(ctx)
	if err != nil {
		depths = map[string]int64{}
	}

	var delivered, failed, pending int64
	delivered, _ = h.db.CountByStatus(ctx, "delivered")
	failed, _ = h.db.CountByStatus(ctx, "failed")
	pending, _ = h.db.CountByStatus(ctx, "pending")

	metrics := map[string]interface{}{
		"queue_depth": map[string]int64{
			"sms":   depths["sms"],
			"email": depths["email"],
			"push":  depths["push"],
		},
		"notifications": map[string]int64{
			"delivered": delivered,
			"failed":    failed,
			"pending":   pending,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
