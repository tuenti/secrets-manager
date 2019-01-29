package backend

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
)

const (
	fakeVaultAddress     = "https://vault.example.com:8200"
	fakeVaultVersion     = "0.11.1"
	fakeVaultEngine      = "kv2"
	fakeVaultClusterID   = "vault-fake-1"
	fakeVaultClusterName = "vault-fake"
)

func TestUpdateMaxTokenTTL(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	maxTokenTTL.Reset()
	metrics.updateVaultMaxTokenTTLMetric(600)
	metricMaxTokenTTL, _ := maxTokenTTL.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName)

	assert.Equal(t, 600.0, testutil.ToFloat64(metricMaxTokenTTL))
}

func TestUpdateTokenTTL(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenTTL.Reset()
	metrics.updateVaultTokenTTLMetric(300)
	metricTokenTTL, _ := tokenTTL.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName)

	assert.Equal(t, 300.0, testutil.ToFloat64(metricTokenTTL))
}

func TestUpdateTokenLookupErrorsTotal(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenRenewalErrorsTotal.Reset()
	metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultLookupSelfOperationName, errors.UnknownErrorType)
	metricTokenRenewalErrorsTotal, _ := tokenRenewalErrorsTotal.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, vaultLookupSelfOperationName, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenRenewalErrorsTotal))
}

func TestUpdateTokenRenewErrorsTotal(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenRenewalErrorsTotal.Reset()
	metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultRenewSelfOperationName, errors.UnknownErrorType)
	metricTokenRenewalErrorsTotal, _ := tokenRenewalErrorsTotal.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, vaultRenewSelfOperationName, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenRenewalErrorsTotal))

	tokenRenewalErrorsTotal.Reset()
	metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultIsRenewableOperationName, errors.VaultTokenNotRenewableErrorType)
	metricTokenRenewalErrorsTotal, _ = tokenRenewalErrorsTotal.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, vaultIsRenewableOperationName, errors.VaultTokenNotRenewableErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenRenewalErrorsTotal))
}

func TestUpdateReadSecretErrorsTotal(t *testing.T) {
	path := "/path/to/secret"
	key := "key"

	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	secretReadErrorsTotal.Reset()
	metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.UnknownErrorType)
	metricSecretReadErrorsTotal, _ := secretReadErrorsTotal.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, path, key, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))

	secretReadErrorsTotal.Reset()
	metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
	metricSecretReadErrorsTotal, _ = secretReadErrorsTotal.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, path, key, errors.BackendSecretNotFoundErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))
}
