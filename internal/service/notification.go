package service

import (
	"context"
	"errors"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("notification not found")
	ErrDuplicateSend = errors.New("duplicate notification (idempotency key already used)")
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Notification, error)
	List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error)
	CancelPending(ctx context.Context, id uuid.UUID) (bool, error)
}

type Enqueuer interface {
	Enqueue(ctx context.Context, notificationID uuid.UUID, channel, priority string) error
}

type NotificationService struct {
	repo    NotificationRepository
	enqueue Enqueuer
}

func NewNotificationService(repo NotificationRepository, enqueue Enqueuer) *NotificationService {
	return &NotificationService{repo: repo, enqueue: enqueue}
}

func (s *NotificationService) Create(ctx context.Context, req domain.CreateNotificationRequest, batchID *uuid.UUID) (*domain.Notification, error) {
	if req.IdempotencyKey != nil {
		existing, err := s.repo.GetByIdempotencyKey(ctx, *req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, ErrDuplicateSend
		}
	}

	n := &domain.Notification{
		ID:         uuid.New(),
		BatchID:    batchID,
		Recipient:  req.Recipient,
		Channel:    req.Channel,
		Content:    req.Content,
		Priority:   req.Priority,
		Status:     domain.StatusPending,
		RetryCount: 0,
		MaxRetries: 3,
	}
	if req.Priority == "" {
		n.Priority = domain.PriorityNormal
	}
	if req.IdempotencyKey != nil {
		n.IdempotencyKey = req.IdempotencyKey
	}
	if req.ScheduledAt != nil {
		n.ScheduledAt = req.ScheduledAt
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	if s.enqueue != nil {
		_ = s.enqueue.Enqueue(ctx, n.ID, string(n.Channel), string(n.Priority))
	}
	return n, nil
}

func (s *NotificationService) CreateBatch(ctx context.Context, requests []domain.CreateNotificationRequest, batchID *uuid.UUID) ([]*domain.Notification, error) {
	if len(requests) > 1000 {
		return nil, errors.New("batch size exceeds 1000")
	}
	if batchID == nil {
		id := uuid.New()
		batchID = &id
	}

	result := make([]*domain.Notification, 0, len(requests))
	for _, req := range requests {
		n, err := s.Create(ctx, req, batchID)
		if err == ErrDuplicateSend {
			continue
		}
		if err != nil {
			return result, err
		}
		result = append(result, n)
	}
	return result, nil
}

func (s *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, ErrNotFound
	}
	return n, nil
}

func (s *NotificationService) GetByBatchID(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return s.repo.GetByBatchID(ctx, batchID)
}

func (s *NotificationService) List(ctx context.Context, f domain.ListFilter) ([]*domain.Notification, error) {
	return s.repo.List(ctx, f)
}

func (s *NotificationService) Cancel(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.CancelPending(ctx, id)
}
