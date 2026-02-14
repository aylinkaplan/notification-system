package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMemoryQueue_PushPop(t *testing.T) {
	ctx := context.Background()
	q := NewMemoryQueue()

	id := uuid.New()
	err := q.Push(ctx, "sms", id, "high")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	dep, _ := q.Depth(ctx, "sms")
	if dep != 1 {
		t.Errorf("Depth(sms) = %d, want 1", dep)
	}

	item, err := q.Pop(ctx, "sms", time.Second)
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}
	if item == nil {
		t.Fatal("Pop returned nil")
	}
	if item.NotificationID != id {
		t.Errorf("NotificationID = %s, want %s", item.NotificationID, id)
	}

	dep2, _ := q.Depth(ctx, "sms")
	if dep2 != 0 {
		t.Errorf("Depth(sms) after Pop = %d, want 0", dep2)
	}
}

func TestMemoryQueue_AllDepths(t *testing.T) {
	ctx := context.Background()
	q := NewMemoryQueue()

	q.Push(ctx, "sms", uuid.New(), "normal")
	q.Push(ctx, "sms", uuid.New(), "normal")
	q.Push(ctx, "email", uuid.New(), "normal")

	depths, err := q.AllDepths(ctx)
	if err != nil {
		t.Fatalf("AllDepths failed: %v", err)
	}
	if depths["sms"] != 2 {
		t.Errorf("depths[sms] = %d, want 2", depths["sms"])
	}
	if depths["email"] != 1 {
		t.Errorf("depths[email] = %d, want 1", depths["email"])
	}
}
