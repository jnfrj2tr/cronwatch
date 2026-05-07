package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/cronwatch/internal/monitor"
)

func newExportMetrics(t *testing.T) *monitor.MetricsStore {
	t.Helper()
	ms := monitor.NewMetricsStore([]string{"backup", "cleanup"})
	ms.RecordHeartbeat("backup")
	ms.RecordHeartbeat("backup")
	ms.RecordMiss("backup")
	ms.RecordMiss("cleanup")
	return ms
}

func TestMetricsExport_PrometheusFormat(t *testing.T) {
	ms := newExportMetrics(t)
	h := newMetricsExportHandler(ms, []string{"backup", "cleanup"})

	req := httptest.NewRequest(http.MethodGet, "/metrics/export", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()

	if !strings.Contains(body, "cronwatch_heartbeats_total{job=\"backup\"} 2") {
		t.Errorf("missing backup heartbeat line; body:\n%s", body)
	}
	if !strings.Contains(body, "cronwatch_misses_total{job=\"backup\"} 1") {
		t.Errorf("missing backup miss line; body:\n%s", body)
	}
	if !strings.Contains(body, "cronwatch_misses_total{job=\"cleanup\"} 1") {
		t.Errorf("missing cleanup miss line; body:\n%s", body)
	}
	if !strings.Contains(body, "cronwatch_uptime_ratio{job=\"backup\"}") {
		t.Errorf("missing uptime ratio line; body:\n%s", body)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected content-type: %s", ct)
	}
}

func TestMetricsExport_MethodNotAllowed(t *testing.T) {
	ms := monitor.NewMetricsStore([]string{"backup"})
	h := newMetricsExportHandler(ms, []string{"backup"})

	req := httptest.NewRequest(http.MethodPost, "/metrics/export", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestMetricsExport_UnknownJobSkipped(t *testing.T) {
	ms := monitor.NewMetricsStore([]string{"backup"})
	h := newMetricsExportHandler(ms, []string{"backup", "ghost"})

	req := httptest.NewRequest(http.MethodGet, "/metrics/export", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if strings.Contains(body, "ghost") {
		t.Errorf("ghost job should be skipped; body:\n%s", body)
	}
}
