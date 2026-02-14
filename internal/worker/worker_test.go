package worker

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	tests := []struct {
		retry int
		want  time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // capped
		{10, 30 * time.Second},
	}
	for _, tt := range tests {
		got := Backoff(tt.retry)
		if got != tt.want {
			t.Errorf("Backoff(%d) = %v, want %v", tt.retry, got, tt.want)
		}
	}
}
