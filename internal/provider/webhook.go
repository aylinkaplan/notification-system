package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type WebhookProvider struct {
	baseURL    string
	httpClient *http.Client
}

func NewWebhookProvider(baseURL string) *WebhookProvider {
	return &WebhookProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SendRequest struct {
	To      string `json:"to"`
	Channel string `json:"channel"`
	Content string `json:"content"`
}

type SendResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func (p *WebhookProvider) Send(ctx context.Context, webhookUUID, to, channel, content string) (*SendResponse, error) {
	url := fmt.Sprintf("%s/%s", p.baseURL, webhookUUID)
	reqBody := SendRequest{To: to, Channel: channel, Content: content}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var result SendResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// 202 received but body is not valid JSON (e.g. webhook.site default page)
		result = SendResponse{
			MessageID: uuid.New().String(),
			Status:    "accepted",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}
	if result.MessageID == "" {
		result.MessageID = uuid.New().String()
	}
	return &result, nil
}
