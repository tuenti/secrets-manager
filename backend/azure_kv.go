package backend

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/go-autorest/autorest"
	"github.com/go-logr/logr"
	"github.com/tuenti/secrets-manager/errors"
)

var akvMetrics *azureKVMetrics

const (
	azureKVEndpoint = "vault.azure.net"
)

type azureKVClient struct {
	client       *azsecrets.Client
	keyvaultName string
	context      context.Context
	logger       logr.Logger
}

func azureKeyVaultClient(ctx context.Context, l logr.Logger, cfg Config) (*azureKVClient, error) {
	logger := l.WithName("azure-kv").WithValues(
		"azure_kv_name", cfg.AzureKVName,
		"azure_kv_tenant", cfg.AzureKVTenantID)

	opts := azidentity.ManagedIdentityCredentialOptions{}
	if cfg.AzureKVManagedClientID != "" {
		opts.ID = azidentity.ClientID(cfg.AzureKVManagedClientID)
	} else if cfg.AzureKVManagedResourceID != "" {
		opts.ID = azidentity.ResourceID(cfg.AzureKVManagedResourceID)
	}

	managed, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		logger.Error(err, "Error occured while authenticating using Azure managed identity")
	}

	spEnv, err := azidentity.NewEnvironmentCredential(nil)
	if err != nil {
		logger.Error(err, "Error occured while authenticating using Azure Service Principal with environment variables")
	}

	spSecret, err := azidentity.NewClientSecretCredential(cfg.AzureKVTenantID, cfg.AzureKVClientID, cfg.AzureKVClientSecret, nil)
	if err != nil {
		logger.Error(err, "Error occured while authenticating using Azure Service Principal")
	}

	cred, err := azidentity.NewChainedTokenCredential([]azcore.TokenCredential{managed, spEnv, spSecret}, nil)
	if err != nil {
		logger.Error(err, "Error occured while authenticating to Azure")
	}
	akvMetrics = newAzureKVMetrics(cfg.AzureKVName, cfg.AzureKVTenantID)
	vaultEndpoint := fmt.Sprintf("https://%s.%s", cfg.AzureKVName, azureKVEndpoint)
	akvClient, err := azsecrets.NewClient(vaultEndpoint, cred, nil)

	if err != nil {
		logger.Error(err, "Error occured while creating Azure KV client")
		akvMetrics.updateLoginErrorsTotalMetric()
	}

	logger.Info("successfully logged into Azure KeyVault")

	client := azureKVClient{
		client:       akvClient,
		keyvaultName: cfg.AzureKVName,
		context:      ctx,
		logger:       logger,
	}

	return &client, err
}

func (c *azureKVClient) ReadSecret(path string, key string) (string, error) {
	data := ""

	// TODO: Add support for secret version?
	result, err := c.client.GetSecret(c.context, path, nil)

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
