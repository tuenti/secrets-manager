package secretsmanager

import(
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometeheus metrics: https://prometheus.io
	secretSyncErrorsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "sync_errors_count",
		Help:      "The count errors when trying to sync a secret from backend into K8s",
	}, []string{"name", "namespace"})

	secretLastUpdated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "last_updated",
		Help:      "The last update timestamp as a Unix time (the number of seconds elapsed since January 1, 1970 UTC)",
	}, []string{"name", "namespace"})
)

func init() {
	prometheus.MustRegister(secretSyncErrorsCount)
	prometheus.MustRegister(secretLastUpdated)
}
