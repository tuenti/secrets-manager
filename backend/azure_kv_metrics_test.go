package backend

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
)

const (
	fakeKeyVaultName   = "azure-keyvault-fake-name"
	fakeKeyVaultTenant = "01234567-0123-0123-0123-0123456789ab"
)

func TestAzureKVUpdateLoginErrorsTotal(t *testing.T) {
	metrics := newAzureKVMetrics(fakeKeyVaultName, fakeKeyVaultTenant)
	azureKVLoginErrorsTotal.Reset()
	metrics.updateLoginErrorsTotalMetric()
	metricLoginErrors, _ := azureKVLoginErrorsTotal.GetMetricWithLabelValues(fakeKeyVaultName, fakeKeyVaultTenant)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricLoginErrors))
}

func TestAzureKVUpdateReadSecretErrorsTotal(t *testing.T) {
	path := "/path/to/secret"
	key := ""

	metrics := newAzureKVMetrics(fakeKeyVaultName, fakeKeyVaultTenant)
	azureKVSecretReadErrorsTotal.Reset()
	metrics.updateSecretReadErrorsTotalMetric(path, errors.UnknownErrorType)
	metricSecretReadErrorsTotal, _ := azureKVSecretReadErrorsTotal.GetMetricWithLabelValues(fakeKeyVaultName, fakeKeyVaultTenant, path, key, errors.UnknownErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))

	azureKVSecretReadErrorsTotal.Reset()
	metrics.updateSecretReadErrorsTotalMetric(path, errors.BackendSecretNotFoundErrorType)
	metricSecretReadErrorsTotal, _ = azureKVSecretReadErrorsTotal.GetMetricWithLabelValues(fakeKeyVaultName, fakeKeyVaultTenant, path, key, errors.BackendSecretNotFoundErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))

	azureKVSecretReadErrorsTotal.Reset()
	metrics.updateSecretReadErrorsTotalMetric(path, errors.BackendSecretForbiddenErrorType)
	metricSecretReadErrorsTotal, _ = azureKVSecretReadErrorsTotal.GetMetricWithLabelValues(fakeKeyVaultName, fakeKeyVaultTenant, path, key, errors.BackendSecretForbiddenErrorType)

	assert.Equal(t, 1.0, testutil.ToFloat64(metricSecretReadErrorsTotal))
}
