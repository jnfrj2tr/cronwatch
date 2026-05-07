package api

import (
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

// RegisterRoutes wires all API handlers onto the given mux.
func RegisterRoutes(
	mux *http.ServeMux,
	m *monitor.Monitor,
	silences *monitor.SilenceStore,
	history *monitor.History,
	metrics *monitor.MetricsStore,
	tags *monitor.TagStore,
	annotations *monitor.AnnotationStore,
	deps *monitor.DependencyStore,
) {
	h := NewHandler(m)
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/heartbeat", h.handleHeartbeat)
	mux.HandleFunc("/status", h.handleStatus)

	sh := &silenceHandler{silences: silences, monitor: m}
	mux.HandleFunc("/silence", sh.handleSilence)
	mux.HandleFunc("/unsilence", sh.handleUnsilence)

	hh := &historyHandler{history: history, monitor: m}
	mux.HandleFunc("/history", hh.handleHistory)

	mh := newMetricsHandler(metrics, m)
	mux.HandleFunc("/metrics", mh.handleMetrics)

	eh := newMetricsExportHandler(metrics, m)
	mux.HandleFunc("/metrics/export", eh.handleExport)

	th := newTagsHandler(tags, m)
	mux.HandleFunc("/tags", th.ServeHTTP)

	ah := newAnnotationsHandler(annotations, m)
	mux.HandleFunc("/annotations", ah.ServeHTTP)

	dh := newDependenciesHandler(deps)
	mux.HandleFunc("/dependencies", dh.ServeHTTP)
	mux.HandleFunc("/dependencies/all", dh.handleAll)
}
