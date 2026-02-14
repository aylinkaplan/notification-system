package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QueuedItem is a single item in the queue.
type QueuedItem struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Priority       string    `json:"priority"`
	QueuedAt       time.Time `json:"queued_at"`
}

// Queue is the interface for the notification queue (e.g. RabbitMQ).
type Queue interface {
	Push(ctx context.Context, channel string, notificationID uuid.UUID, priority string) error
	Pop(ctx context.Context, channel string, timeout time.Duration) (*QueuedItem, error)
	Depth(ctx context.Context, channel string) (int64, error)
	AllDepths(ctx context.Context) (map[string]int64, error)
}
