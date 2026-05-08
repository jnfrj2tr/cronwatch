package api

import (
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

// RegisterRoutes wires all API handlers onto the given mux.
func RegisterRoutes(
	mux *http.ServeMux,
	h *Handler,
	silenceStore *monitor.SilenceStore,
	history *monitor.History,
	metrics *monitor.MetricsStore,
	runLog *monitor.RunLog,
	retries *monitor.RetryStore,
	tags *monitor.TagStore,
	annotations *monitor.AnnotationStore,
	deps *monitor.DependencyStore,
	maintenance *monitor.MaintenanceStore,
) {
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/heartbeat", h.HandleHeartbeat)
	mux.HandleFunc("/status", h.HandleStatus)

	mux.Handle("/silence", newSilenceHandler(silenceStore))
	mux.Handle("/history", newHistoryHandler(history))
	mux.Handle("/metrics", newMetricsHandler(metrics))
	mux.Handle("/metrics/export", newMetricsExportHandler(metrics))
	mux.Handle("/runlog", newRunLogHandler(runLog))
	mux.Handle("/retries", newRetriesHandler(retries))
	mux.Handle("/retries/all", newRetriesAllHandler(retries))
	mux.Handle("/tags", newTagsHandler(tags))
	mux.Handle("/annotations", newAnnotationsHandler(annotations))
	mux.Handle("/dependencies", newDependenciesHandler(deps))
	mux.Handle("/maintenance", newMaintenanceHandler(maintenance))
}
