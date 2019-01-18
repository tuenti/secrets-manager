package kubernetes

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometeheus metrics: https://prometheus.io
	secretReadErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "k8s",
		Name:      "secret_read_error_count",
		Help:      "Errors count when reading a secret from Kubernetes",
	}, []string{"name", "namespace"})
	secretUpdateErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "k8s",
		Name:      "secret_update_error_count",
		Help:      "Error count when updating (and also creating) a secret in Kubernetes",
	}, []string{"name", "namespace"})
)

func init() {
	prometheus.MustRegister(secretReadErrorCount)
	prometheus.MustRegister(secretUpdateErrorCount)
}
