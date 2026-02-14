package queue

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryQueue is an in-memory implementation of Queue for testing.
type MemoryQueue struct {
	mu       sync.Mutex
	channels map[string][]*QueuedItem
}

// NewMemoryQueue returns a new in-memory queue.
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		channels: make(map[string][]*QueuedItem),
	}
}

func (m *MemoryQueue) Push(ctx context.Context, channel string, notificationID uuid.UUID, priority string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.channels[channel] == nil {
		m.channels[channel] = make([]*QueuedItem, 0)
	}
	m.channels[channel] = append(m.channels[channel], &QueuedItem{
		NotificationID: notificationID,
		Priority:       priority,
		QueuedAt:       time.Now(),
	})
	return nil
}

func (m *MemoryQueue) Pop(ctx context.Context, channel string, timeout time.Duration) (*QueuedItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.channels[channel]
	if len(list) == 0 {
		return nil, nil
	}
	item := list[0]
	m.channels[channel] = list[1:]
	return item, nil
}

func (m *MemoryQueue) Depth(ctx context.Context, channel string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.channels[channel])), nil
}

func (m *MemoryQueue) AllDepths(ctx context.Context) (map[string]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int64{"sms": 0, "email": 0, "push": 0}
	for ch, list := range m.channels {
		out[ch] = int64(len(list))
	}
	return out, nil
}
