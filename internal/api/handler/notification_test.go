package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/aylinkaplan/notification-system/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type handlerMockRepo struct {
	byID     map[uuid.UUID]*domain.Notification
	byBatch  map[uuid.UUID][]*domain.Notification
	byKey    map[string]*domain.Notification
	list     []*domain.Notification
	create   func(context.Context, *domain.Notification) error
	cancelOK bool
}

func (m *handlerMockRepo) Create(ctx context.Context, n *domain.Notification) error {
	if m.create != nil {
		return m.create(ctx, n)
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*domain.Notification)
	}
	m.byID[n.ID] = n
	return nil
}

func (m *handlerMockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return m.byID[id], nil
}

func (m *handlerMockRepo) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return m.byBatch[batchID], nil
}

func (m *handlerMockRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Notification, error) {
	return m.byKey[key], nil
}

func (m *handlerMockRepo) List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error) {
	return m.list, nil
}

func (m *handlerMockRepo) CancelPending(ctx context.Context, id uuid.UUID) (bool, error) {
	return m.cancelOK, nil
}

type handlerMockEnqueuer struct{}

func (m *handlerMockEnqueuer) Enqueue(ctx context.Context, notificationID uuid.UUID, channel, priority string) error {
	return nil
}

func TestNotificationHandler_Create_Success(t *testing.T) {
	repo := &handlerMockRepo{}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	body := `{"recipient":"+905551234567","channel":"sms","content":"Hello","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/notifications", h.Create)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	var out domain.Notification
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if out.Status != domain.StatusPending {
		t.Errorf("status = %s, want pending", out.Status)
	}
}

func TestNotificationHandler_Create_Validation(t *testing.T) {
	repo := &handlerMockRepo{}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	tests := []struct {
		name string
		body string
		want int
	}{
		{"missing recipient", `{"channel":"sms","content":"x"}`, 400},
		{"invalid channel", `{"recipient":"a","channel":"invalid","content":"x"}`, 400},
		{"invalid priority", `{"recipient":"a","channel":"sms","content":"x","priority":"invalid"}`, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r := chi.NewRouter()
			r.Post("/notifications", h.Create)
			r.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestNotificationHandler_GetByID(t *testing.T) {
	id := uuid.New()
	n := &domain.Notification{ID: id, Recipient: "a", Channel: domain.ChannelSMS, Content: "hi", Status: domain.StatusPending}
	repo := &handlerMockRepo{byID: map[uuid.UUID]*domain.Notification{id: n}}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/notifications/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/notifications/{id}", h.GetByID)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNotificationHandler_GetByID_NotFound(t *testing.T) {
	repo := &handlerMockRepo{byID: map[uuid.UUID]*domain.Notification{}}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/notifications/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/notifications/{id}", h.GetByID)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestNotificationHandler_CreateBatch_TooMany(t *testing.T) {
	repo := &handlerMockRepo{}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	notifs := make([]domain.CreateNotificationRequest, 1001)
	for i := range notifs {
		notifs[i] = domain.CreateNotificationRequest{Recipient: "a", Channel: domain.ChannelSMS, Content: "x"}
	}
	body, _ := json.Marshal(map[string]interface{}{"notifications": notifs})
	req := httptest.NewRequest(http.MethodPost, "/notifications/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/notifications/batch", h.CreateBatch)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestNotificationHandler_Cancel(t *testing.T) {
	id := uuid.New()
	repo := &handlerMockRepo{byID: map[uuid.UUID]*domain.Notification{id: {}}, cancelOK: true}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/notifications/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/notifications/{id}", h.Cancel)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNotificationHandler_Cancel_Conflict(t *testing.T) {
	id := uuid.New()
	repo := &handlerMockRepo{byID: map[uuid.UUID]*domain.Notification{id: {}}, cancelOK: false}
	svc := service.NewNotificationService(repo, &handlerMockEnqueuer{})
	h := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/notifications/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Delete("/notifications/{id}", h.Cancel)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}
