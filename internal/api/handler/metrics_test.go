package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aylinkaplan/notification-system/internal/queue"
)

// fakeCountReader is a test double for NotificationCountReader.
type fakeCountReader struct {
	delivered, failed, pending int64
}

func (f *fakeCountReader) CountByStatus(ctx context.Context, status string) (int64, error) {
	switch status {
	case "delivered":
		return f.delivered, nil
	case "failed":
		return f.failed, nil
	case "pending":
		return f.pending, nil
	}
	return 0, nil
}

func TestMetricsHandler_ServeHTTP(t *testing.T) {
	db := &fakeCountReader{delivered: 10, failed: 1, pending: 2}
	q := queue.NewMemoryQueue()
	h := NewMetricsHandler(db, q)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s", rec.Header().Get("Content-Type"))
	}
	// Decode and assert counts
	var out struct {
		Notifications struct {
			Delivered int64 `json:"delivered"`
			Failed    int64 `json:"failed"`
			Pending   int64 `json:"pending"`
		} `json:"notifications"`
	}
	if err := jsonDecode(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Notifications.Delivered != 10 {
		t.Errorf("delivered = %d, want 10", out.Notifications.Delivered)
	}
	if out.Notifications.Failed != 1 {
		t.Errorf("failed = %d, want 1", out.Notifications.Failed)
	}
	if out.Notifications.Pending != 2 {
		t.Errorf("pending = %d, want 2", out.Notifications.Pending)
	}
}

// jsonDecode decodes JSON into v (test helper to avoid importing encoding/json in test for type assertion).
func jsonDecode(b []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewReader(b)).Decode(v)
}
