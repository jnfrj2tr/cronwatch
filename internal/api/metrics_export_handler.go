package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/user/cronwatch/internal/monitor"
)

type metricsExportHandler struct {
	metrics *monitor.MetricsStore
	jobs    []string
}

func newMetricsExportHandler(metrics *monitor.MetricsStore, jobs []string) *metricsExportHandler {
	return &metricsExportHandler{metrics: metrics, jobs: jobs}
}

// ServeHTTP writes Prometheus-compatible text exposition format.
func (h *metricsExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sb strings.Builder

	all := h.metrics.All()

	sb.WriteString("# HELP cronwatch_heartbeats_total Total number of heartbeats received per job.\n")
	sb.WriteString("# TYPE cronwatch_heartbeats_total counter\n")
	for _, job := range h.jobs {
		m, ok := all[job]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("cronwatch_heartbeats_total{job=%q} %d\n", job, m.Heartbeats))
	}

	sb.WriteString("# HELP cronwatch_misses_total Total number of missed runs per job.\n")
	sb.WriteString("# TYPE cronwatch_misses_total counter\n")
	for _, job := range h.jobs {
		m, ok := all[job]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("cronwatch_misses_total{job=%q} %d\n", job, m.Misses))
	}

	sb.WriteString("# HELP cronwatch_uptime_ratio Ratio of successful runs to total expected runs per job.\n")
	sb.WriteString("# TYPE cronwatch_uptime_ratio gauge\n")
	for _, job := range h.jobs {
		m, ok := all[job]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("cronwatch_uptime_ratio{job=%q} %.4f\n", job, m.UptimeRatio))
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}
