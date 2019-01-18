package backend

import "github.com/prometheus/client_golang/prometheus"

const (
	vaultTokenExpired    = 1
	vaultTokenNotExpired = 0
)

var (
	vaultLabelNames  = []string{"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"}
	secretLabelNames = []string{"path", "key", "error"}

	// Prometeheus metrics: https://prometheus.io
	tokenExpired = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "token_expired",
		Help:      "The state of the token: 1 = expired; 0 = still valid",
	}, vaultLabelNames)
	tokenTTL = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "token_ttl",
		Help:      "Vault token TTL",
	}, vaultLabelNames)
	secretReadErrorsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "read_secret_errors_count",
		Help:      "Vault read operations counter",
	}, append(vaultLabelNames, secretLabelNames...))
)

type vaultMetrics struct {
	vaultLabels map[string]string
}

func init() {
	prometheus.MustRegister(tokenExpired)
	prometheus.MustRegister(tokenTTL)
	prometheus.MustRegister(secretReadErrorsCount)
}

func newVaultMetrics(vaultAddr string, vaultVersion string, vaultEngine string, vaultClusterId string, vaultClusterName string) *vaultMetrics {
	labels := make(map[string]string, len(vaultLabelNames))
	labels["vault_addr"] = vaultAddr
	labels["vault_engine"] = vaultEngine
	labels["vault_version"] = vaultVersion
	labels["vault_cluster_id"] = vaultClusterId
	labels["vault_cluster_name"] = vaultClusterName

	return &vaultMetrics{vaultLabels: labels}
}

func (vm *vaultMetrics) updateVaultTokenExpiredMetric(value int) {
	if value != vaultTokenExpired && value != vaultTokenNotExpired {
		logger.Errorf("refusing to update secrets_manager_vault_token_expired metric with value %d. Allowed values are %d and %d", value, vaultTokenExpired, vaultTokenNotExpired)
		return
	}

	tokenExpired.WithLabelValues(
		vm.vaultLabels["vault_addr"],
		vm.vaultLabels["vault_engine"],
		vm.vaultLabels["vault_version"],
		vm.vaultLabels["vault_cluster_id"],
		vm.vaultLabels["vault_cluster_name"]).Set(float64(value))
}

func (vm *vaultMetrics) updateVaultTokenTTLMetric(value int64) {
	tokenTTL.WithLabelValues(
		vm.vaultLabels["vault_addr"],
		vm.vaultLabels["vault_engine"],
		vm.vaultLabels["vault_version"],
		vm.vaultLabels["vault_cluster_id"],
		vm.vaultLabels["vault_cluster_name"]).Set(float64(value))
}

func (vm *vaultMetrics) updateVaultSecretReadErrorsCountMetric(path string, key string, errorType string) {

	secretReadErrorsCount.WithLabelValues(
		vm.vaultLabels["vault_addr"],
		vm.vaultLabels["vault_engine"],
		vm.vaultLabels["vault_version"],
		vm.vaultLabels["vault_cluster_id"],
		vm.vaultLabels["vault_cluster_name"],
		path,
		key,
		errorType).Inc()
}
