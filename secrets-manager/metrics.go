package secretsmanager

import(
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometeheus metrics: https://prometheus.io
	secretUpdated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "updated",
		Help:      "The up-to-date state of the secret: 1 = succesfulyy updated; 0 = couldn't update it",
	}, []string{"name", "namespace"})

	secretLastUpdated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "secret",
		Name:      "last_updated",
		Help:      "The last update timestamp as a Unix time (the number of seconds elapsed since January 1, 1970 UTC)",
	}, []string{"name", "namespace"})
)

func init() {
	prometheus.MustRegister(secretUpdated)
	prometheus.MustRegister(secretLastUpdated)
}
