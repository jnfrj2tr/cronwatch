package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := alert.NewWebhookNotifier(srv.URL)
	a := alert.Alert{
		JobName:    "deploy",
		Level:      alert.LevelError,
		Message:    "missed",
		OccurredAt: time.Now(),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["job_name"] != "deploy" {
		t.Errorf("expected job_name=deploy, got %q", received["job_name"])
	}
	if received["level"] != "ERROR" {
		t.Errorf("expected level=ERROR, got %q", received["level"])
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := alert.NewWebhookNotifier(srv.URL)
	a := alert.Alert{JobName: "job", Level: alert.LevelWarn, Message: "late", OccurredAt: time.Now()}
	if err := n.Send(a); err == nil {
		t.Error("expected error for non-2xx status")
	}
}

func TestWebhookNotifier_Send_BadURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/nope")
	a := alert.Alert{JobName: "job", Level: alert.LevelError, Message: "x", OccurredAt: time.Now()}
	if err := n.Send(a); err == nil {
		t.Error("expected error for unreachable URL")
	}
}
