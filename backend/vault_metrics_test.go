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

func TestUpdateTokenExpired(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenExpired.Reset()
	metrics.updateVaultTokenExpiredMetric(1)
	metricTokenExpired, _ := tokenExpired.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenExpired))
}

func TestUpdateTokenTTL(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenTTL.Reset()
	metrics.updateVaultTokenTTLMetric(300)
	metricTokenTTL, _ := tokenTTL.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName)

	assert.Equal(t, 300.0, testutil.ToFloat64(metricTokenTTL))
}

func TestUpdateTokenLookupErrorsCount(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenLookupErrorsCount.Reset()
	metrics.updateVaultTokenLookupErrorsCountMetric(errors.UnknownErrorType)
	metricTokenLookupErrorsCount, _ := tokenLookupErrorsCount.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenLookupErrorsCount))
}

func TestUpdateTokenRenewErrorsCount(t *testing.T) {
	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	tokenRenewErrorsCount.Reset()
	metrics.updateVaultTokenRenewErrorsCountMetric(errors.UnknownErrorType)
	metricTokenRenewErrorsCount, _ := tokenRenewErrorsCount.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenRenewErrorsCount))

	tokenRenewErrorsCount.Reset()
	metrics.updateVaultTokenRenewErrorsCountMetric(errors.VaultTokenNotRenewableErrorType)
	metricTokenRenewErrorsCount, _ = tokenRenewErrorsCount.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, errors.VaultTokenNotRenewableErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricTokenRenewErrorsCount))
}

func TestUpdateReadSecretErrorsCount(t *testing.T) {
	path := "/path/to/secret"
	key := "key"

	metrics := newVaultMetrics(fakeVaultAddress, fakeVaultVersion, fakeVaultEngine, fakeVaultClusterID, fakeVaultClusterName)
	secretReadErrorsCount.Reset()
	metrics.updateVaultSecretReadErrorsCountMetric(path, key, errors.UnknownErrorType)
	metricSecretReadErrorsCount, _ := secretReadErrorsCount.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, path, key, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsCount))

	secretReadErrorsCount.Reset()
	metrics.updateVaultSecretReadErrorsCountMetric(path, key, errors.BackendSecretNotFoundErrorType)
	metricSecretReadErrorsCount, _ = secretReadErrorsCount.GetMetricWithLabelValues(fakeVaultAddress, fakeVaultEngine, fakeVaultVersion, fakeVaultClusterID, fakeVaultClusterName, path, key, errors.BackendSecretNotFoundErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsCount))
}
