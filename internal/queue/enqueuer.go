package queue

import (
	"context"

	"github.com/google/uuid"
)

type EnqueuerAdapter struct {
	q Queue
}

func NewEnqueuerAdapter(q Queue) *EnqueuerAdapter {
	return &EnqueuerAdapter{q: q}
}

func (e *EnqueuerAdapter) Enqueue(ctx context.Context, notificationID uuid.UUID, channel, priority string) error {
	return e.q.Push(ctx, channel, notificationID, priority)
}
