package backend

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	azureKVLabelNames            = []string{"azure_kv_name", "azure_kv_tenant"}
	azureKVSecretReadErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "azure_kv",
		Name:      "read_secret_errors_total",
		Help:      "AzureKV read operations counter",
	}, append(azureKVLabelNames, secretLabelNames...))
	azureKVLoginErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "azure_kv",
		Name:      "login_errors_total",
		Help:      "AzureKV login errors counter",
	}, azureKVLabelNames)
)

type azureKVMetrics struct {
	labels map[string]string
}

func newAzureKVMetrics(keyvaultName string, tenantID string) *azureKVMetrics {
	labels := make(map[string]string, len(azureKVLabelNames))
	labels["azure_kv_name"] = keyvaultName
	labels["azure_kv_tenant"] = tenantID

	return &azureKVMetrics{labels: labels}
}

func (vm *azureKVMetrics) updateSecretReadErrorsTotalMetric(path string, errorType string) {
	azureKVSecretReadErrorsTotal.WithLabelValues(
		vm.labels["azure_kv_name"],
		vm.labels["azure_kv_tenant"],
		path,
		"",
		errorType,
	).Inc()
}

func (vm *azureKVMetrics) updateLoginErrorsTotalMetric() {
	azureKVLoginErrorsTotal.WithLabelValues(
		vm.labels["azure_kv_name"],
		vm.labels["azure_kv_tenant"],
	).Inc()
}
