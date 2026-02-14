package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aylinkaplan/notification-system/internal/api/handler"
	"github.com/aylinkaplan/notification-system/internal/config"
	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/aylinkaplan/notification-system/internal/queue"
	"github.com/aylinkaplan/notification-system/internal/service"
	"github.com/google/uuid"
)

// apiTestRepo implements service.NotificationRepository for router tests.
type apiTestRepo struct {
	byID map[uuid.UUID]*domain.Notification
}

func (r *apiTestRepo) Create(ctx context.Context, n *domain.Notification) error {
	if r.byID == nil {
		r.byID = make(map[uuid.UUID]*domain.Notification)
	}
	r.byID[n.ID] = n
	return nil
}

func (r *apiTestRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return r.byID[id], nil
}

func (r *apiTestRepo) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}

func (r *apiTestRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Notification, error) {
	return nil, nil
}

func (r *apiTestRepo) List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error) {
	return nil, nil
}

func (r *apiTestRepo) CancelPending(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}

type apiTestEnqueuer struct{}

func (apiTestEnqueuer) Enqueue(ctx context.Context, notificationID uuid.UUID, channel, priority string) error {
	return nil
}

// apiTestCountReader for metrics (implements handler.NotificationCountReader).
type apiTestCountReader struct{}

func (apiTestCountReader) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}

func TestRouter_Health(t *testing.T) {
	cfg := &config.Config{}
	svc := service.NewNotificationService(&apiTestRepo{}, &apiTestEnqueuer{})
	notifHandler := handler.NewNotificationHandler(svc)
	metricsHandler := handler.NewMetricsHandler(&apiTestCountReader{}, queue.NewMemoryQueue())

	// db and q are nil - we must not call /ready in this test
	r := NewRouter(cfg, nil, nil, notifHandler, metricsHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s", rec.Header().Get("Content-Type"))
	}
}

func TestRouter_OpenAPI(t *testing.T) {
	cfg := &config.Config{}
	svc := service.NewNotificationService(&apiTestRepo{}, &apiTestEnqueuer{})
	notifHandler := handler.NewNotificationHandler(svc)
	metricsHandler := handler.NewMetricsHandler(&apiTestCountReader{}, queue.NewMemoryQueue())

	r := NewRouter(cfg, nil, nil, notifHandler, metricsHandler)

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /openapi.yaml status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if body == "" || len(body) < 50 {
		t.Error("openapi.yaml body too short")
	}
}

func TestRouter_Docs(t *testing.T) {
	cfg := &config.Config{}
	svc := service.NewNotificationService(&apiTestRepo{}, &apiTestEnqueuer{})
	notifHandler := handler.NewNotificationHandler(svc)
	metricsHandler := handler.NewMetricsHandler(&apiTestCountReader{}, queue.NewMemoryQueue())

	r := NewRouter(cfg, nil, nil, notifHandler, metricsHandler)

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// 200 (HTML) or 302 (redirect) both ok
	if rec.Code != http.StatusOK && rec.Code != http.StatusMovedPermanently && rec.Code != http.StatusFound {
		t.Errorf("GET /docs status = %d, want 200 or 3xx", rec.Code)
	}
}
