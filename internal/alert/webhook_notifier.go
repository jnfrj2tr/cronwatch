package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends alerts as JSON POST requests to a URL.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a default HTTP client.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type webhookPayload struct {
	JobName    string `json:"job_name"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurred_at"`
}

// Send POSTs the alert as JSON to the configured webhook URL.
func (w *WebhookNotifier) Send(a Alert) error {
	payload := webhookPayload{
		JobName:    a.JobName,
		Level:      string(a.Level),
		Message:    a.Message,
		OccurredAt: a.OccurredAt.Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}
	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
