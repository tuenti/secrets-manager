package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	registry prometheus.Registry
	// Prometeheus metrics: https://prometheus.io
	secretReadErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "secret_read_errors_total",
		Help:      "Errors total count when reading a secret from Kubernetes",
	}, []string{"name", "namespace"})

	secretUpdateErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "secret_update_errors_total",
		Help:      "Error total count when updating (and also creating) a secret in Kubernetes",
	}, []string{"name", "namespace"})

	secretSyncErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "sync_errors_total",
		Help:      "Secrets synchronization total errors.",
	}, []string{"name", "namespace"})

	secretLastUpdated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "last_updated",
		Help:      "The last update timestamp as a Unix time (the number of seconds elapsed since January 1, 1970 UTC)",
	}, []string{"name", "namespace"})

	secretLastSyncStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "last_sync_status",
		Help:      "The result of the last sync of a secret. 1 = OK, 0 = Error",
	}, []string{"name", "namespace"})
)

func init() {
	r := metrics.Registry
	r.MustRegister(secretReadErrorsTotal)
	r.MustRegister(secretUpdateErrorsTotal)
	r.MustRegister(secretSyncErrorsTotal)
	r.MustRegister(secretLastUpdated)
	r.MustRegister(secretLastSyncStatus)
}
