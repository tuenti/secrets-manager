package backend

import "github.com/prometheus/client_golang/prometheus"

const (
	vaultLookupSelfOperationName  = "lookup-self"
	vaultRenewSelfOperationName   = "renew-self"
	vaultIsRenewableOperationName = "is-renewable"
)

var (
	vaultLabelNames      = []string{"vault_address", "vault_engine", "vault_version", "vault_cluster_id", "vault_cluster_name"}
	secretLabelNames     = []string{"path", "key", "error"}
	vaultErrorLabelNames = []string{"vault_operation", "error"}

	// Prometeheus metrics: https://prometheus.io
	tokenTTL = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "token_ttl",
		Help:      "Vault token TTL",
	}, vaultLabelNames)
	maxTokenTTL = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "max_token_ttl",
		Help:      "secrets-manager max Vault token TTL",
	}, vaultLabelNames)
	tokenRenewalErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "token_renewal_errors_total",
		Help:      "Vault token renewal errors counter",
	}, append(vaultLabelNames, vaultErrorLabelNames...))
	secretReadErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "secrets_manager",
		Subsystem: "vault",
		Name:      "read_secret_errors_total",
		Help:      "Vault read operations counter",
	}, append(vaultLabelNames, secretLabelNames...))
)

type vaultMetrics struct {
	vaultLabels map[string]string
}

func init() {
	prometheus.MustRegister(tokenTTL)
	prometheus.MustRegister(maxTokenTTL)
	prometheus.MustRegister(tokenRenewalErrorsTotal)
	prometheus.MustRegister(secretReadErrorsTotal)
}

func newVaultMetrics(vaultAddr string, vaultVersion string, vaultEngine string, vaultClusterID string, vaultClusterName string) *vaultMetrics {
	labels := make(map[string]string, len(vaultLabelNames))
	labels["vault_addr"] = vaultAddr
	labels["vault_engine"] = vaultEngine
	labels["vault_version"] = vaultVersion
	labels["vault_cluster_id"] = vaultClusterID
	labels["vault_cluster_name"] = vaultClusterName

	return &vaultMetrics{vaultLabels: labels}
}

func (vm *vaultMetrics) updateVaultMaxTokenTTLMetric(value int64) {
	maxTokenTTL.WithLabelValues(
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

func (vm *vaultMetrics) updateVaultSecretReadErrorsTotalMetric(path string, key string, errorType string) {
	secretReadErrorsTotal.WithLabelValues(
		vm.vaultLabels["vault_addr"],
		vm.vaultLabels["vault_engine"],
		vm.vaultLabels["vault_version"],
		vm.vaultLabels["vault_cluster_id"],
		vm.vaultLabels["vault_cluster_name"],
		path,
		key,
		errorType).Inc()
}

func (vm *vaultMetrics) updateVaultTokenRenewalErrorsTotalMetric(vaultOperation string, errorType string) {
	tokenRenewalErrorsTotal.WithLabelValues(
		vm.vaultLabels["vault_addr"],
		vm.vaultLabels["vault_engine"],
		vm.vaultLabels["vault_version"],
		vm.vaultLabels["vault_cluster_id"],
		vm.vaultLabels["vault_cluster_name"],
		vaultOperation,
		errorType).Inc()
}
