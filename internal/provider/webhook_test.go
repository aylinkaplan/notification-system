package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookProvider_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(SendResponse{
			MessageID: "msg-123",
			Status:    "accepted",
			Timestamp: "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	prov := NewWebhookProvider(server.URL)
	ctx := context.Background()

	resp, err := prov.Send(ctx, "test-uuid", "+905551234567", "sms", "Hello")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if resp.MessageID != "msg-123" {
		t.Errorf("expected messageId msg-123, got %s", resp.MessageID)
	}
}

func TestWebhookProvider_Send_202NonJSON_StillSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("<html>not json</html>"))
	}))
	defer server.Close()

	prov := NewWebhookProvider(server.URL)
	ctx := context.Background()

	resp, err := prov.Send(ctx, "uuid", "a@b.com", "email", "Hi")
	if err != nil {
		t.Fatalf("Send should succeed on 202 with non-JSON body: %v", err)
	}
	if resp.MessageID == "" {
		t.Error("expected generated messageId when body is not JSON")
	}
}

func TestWebhookProvider_Send_5xx_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	prov := NewWebhookProvider(server.URL)
	ctx := context.Background()

	_, err := prov.Send(ctx, "uuid", "a@b.com", "email", "Hi")
	if err == nil {
		t.Fatal("expected error on 500")
	}
}
