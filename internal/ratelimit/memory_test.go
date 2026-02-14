package ratelimit

import (
	"context"
	"testing"
)

func TestNewMemoryLimiter(t *testing.T) {
	l := NewMemoryLimiter(100)
	if l == nil {
		t.Fatal("NewMemoryLimiter returned nil")
	}
}

func TestMemoryLimiter_Allow(t *testing.T) {
	ctx := context.Background()
	l := NewMemoryLimiter(10) // 10 per second, burst typically 20

	// First allow should succeed
	ok, err := l.Allow(ctx, "key1")
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !ok {
		t.Error("first Allow expected true")
	}

	// Same key uses same limiter
	ok2, _ := l.Allow(ctx, "key1")
	if !ok2 {
		t.Error("second Allow expected true (within burst)")
	}

	// Different key gets its own limiter
	ok3, _ := l.Allow(ctx, "key2")
	if !ok3 {
		t.Error("Allow for key2 expected true")
	}
}

func TestMemoryLimiter_Wait(t *testing.T) {
	ctx := context.Background()
	l := NewMemoryLimiter(100)
	err := l.Wait(ctx, "wait-key")
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}
}
