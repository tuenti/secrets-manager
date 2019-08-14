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
	}, []string{"namespace", "name"})

	secretSyncErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "sync_errors_total",
		Help:      "Secrets synchronization total errors.",
	}, []string{"namespace", "name"})

	secretLastSyncStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "controller",
		Name:      "last_sync_status",
		Help:      "The result of the last sync of a secret. 1 = OK, 0 = Error",
	}, []string{"namespace", "name"})
)

func init() {
	r := metrics.Registry
	r.MustRegister(secretReadErrorsTotal)
	r.MustRegister(secretSyncErrorsTotal)
	r.MustRegister(secretLastSyncStatus)
}
