package backend

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/go-logr/logr"
	"github.com/tuenti/secrets-manager/errors"
)

var akvMetrics *azureKVMetrics

const (
	azureKVEndpoint = "vault.azure.net"
)

type azureKVClient struct {
	client       *keyvault.BaseClient
	keyvaultName string
	tenantID     string
	clientID     string
	clientSecret string
	logger       logr.Logger
}

func azureKeyVaultClient(l logr.Logger, cfg Config) (*azureKVClient, error) {
	logger := l.WithName("azure-kv").WithValues(
		"azure_kv_name", cfg.AzureKVName,
		"azure_kv_tenant", cfg.AzureKVTenantID)

	akvMetrics = newAzureKVMetrics(cfg.AzureKVName, cfg.AzureKVTenantID)
	akvClient := keyvault.New()
	clientCredentialConfig := auth.NewClientCredentialsConfig(cfg.AzureKVClientID, cfg.AzureKVClientSecret, cfg.AzureKVTenantID)

	// From SDK NewClientCredentialsConfig generates a object to azure control plane
	// (By default Resource is set to management.azure.net)
	// There below line was added to access the azure data plane
	// Which is required to access secrets in keyvault

	clientCredentialConfig.Resource = fmt.Sprintf("https://%s", azureKVEndpoint)
	authorizer, err := clientCredentialConfig.Authorizer()

	if err != nil {
		logger.Error(err, "Error occured while creating azure KV authorizer")
		akvMetrics.updateLoginErrorsTotalMetric()
	}
	akvClient.Authorizer = authorizer

	logger.Info("successfully logged into Azure KeyVault")

	client := azureKVClient{
		client:       &akvClient,
		keyvaultName: cfg.AzureKVName,
		tenantID:     cfg.AzureKVTenantID,
		clientID:     cfg.AzureKVClientID,
		clientSecret: cfg.AzureKVClientSecret,
		logger:       logger,
	}

	return &client, err
}

func (c *azureKVClient) ReadSecret(path string, key string) (string, error) {
	data := ""
	uri := fmt.Sprintf("https://%s.%s", c.keyvaultName, azureKVEndpoint)

	// TODO: Add support for secret version?
	result, err := c.client.GetSecret(context.Background(), uri, path, "")

	if err != nil {
		errorType := errors.UnknownErrorType
		if v, ok := err.(autorest.DetailedError); ok {
			if v.StatusCode == 404 {
				errorType = errors.BackendSecretNotFoundErrorType
			}
			if v.StatusCode == 403 {
				errorType = errors.BackendSecretForbiddenErrorType
			}
		}
		akvMetrics.updateSecretReadErrorsTotalMetric(path, errorType)
		return data, err
	}

	data = *result.Value
	return data, err
}
