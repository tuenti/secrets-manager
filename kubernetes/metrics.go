package kubernetes

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometeheus metrics: https://prometheus.io
	secretReadErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "k8s",
		Name:      "secret_read_errors_total",
		Help:      "Errors total count when reading a secret from Kubernetes",
	}, []string{"name", "namespace"})
	secretUpdateErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "k8s",
		Name:      "secret_update_errors_total",
		Help:      "Error total count when updating (and also creating) a secret in Kubernetes",
	}, []string{"name", "namespace"})
)

func init() {
	prometheus.MustRegister(secretReadErrorsTotal)
	prometheus.MustRegister(secretUpdateErrorsTotal)
}
