package service

import (
	"context"
	"testing"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/google/uuid"
)

type mockRepo struct {
	createErr error
	getByID   map[uuid.UUID]*domain.Notification
}

func (m *mockRepo) Create(ctx context.Context, n *domain.Notification) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.getByID == nil {
		m.getByID = make(map[uuid.UUID]*domain.Notification)
	}
	m.getByID[n.ID] = n
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return m.getByID[id], nil
}

func (m *mockRepo) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *mockRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Notification, error) {
	return nil, nil
}

func (m *mockRepo) List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *mockRepo) CancelPending(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}

type mockEnqueuer struct{}

func (m *mockEnqueuer) Enqueue(ctx context.Context, notificationID uuid.UUID, channel, priority string) error {
	return nil
}

func TestNotificationService_Create(t *testing.T) {
	repo := &mockRepo{}
	svc := NewNotificationService(repo, &mockEnqueuer{})
	ctx := context.Background()

	req := domain.CreateNotificationRequest{
		Recipient: "+905551234567",
		Channel:   domain.ChannelSMS,
		Content:   "Hello",
		Priority:  domain.PriorityNormal,
	}

	n, err := svc.Create(ctx, req, nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if n.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if n.Status != domain.StatusPending {
		t.Errorf("expected status pending, got %s", n.Status)
	}
	if n.Recipient != req.Recipient {
		t.Errorf("expected recipient %s, got %s", req.Recipient, n.Recipient)
	}
}

func TestNotificationService_CreateBatch_ExceedsLimit(t *testing.T) {
	repo := &mockRepo{}
	svc := NewNotificationService(repo, &mockEnqueuer{})
	ctx := context.Background()

	reqs := make([]domain.CreateNotificationRequest, 1001)
	for i := range reqs {
		reqs[i] = domain.CreateNotificationRequest{
			Recipient: "+905551234567",
			Channel:   domain.ChannelSMS,
			Content:   "Hi",
		}
	}

	_, err := svc.CreateBatch(ctx, reqs, nil)
	if err == nil {
		t.Error("expected error for batch size > 1000")
	}
}
