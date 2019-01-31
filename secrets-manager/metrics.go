package secretsmanager

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometeheus metrics: https://prometheus.io
	secretSyncErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "sync_errors_total",
		Help:      "Secrets synchronization total errors.",
	}, []string{"name", "namespace"})

	secretLastUpdated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "last_updated",
		Help:      "The last update timestamp as a Unix time (the number of seconds elapsed since January 1, 1970 UTC)",
	}, []string{"name", "namespace"})

	secretLastSyncStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "last_sync_status",
		Help:      "The result of the last sync of a secret. 1 = OK, 0 = Error",
	}, []string{"name", "namespace"})
)

func init() {
	prometheus.MustRegister(secretSyncErrorsTotal)
	prometheus.MustRegister(secretLastUpdated)
	prometheus.MustRegister(secretLastSyncStatus)
}
