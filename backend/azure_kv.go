package backend

import (
	"context"
	goerrors "errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
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

// getAzureCredential finds the better way to authenticate to Azure
func getAzureCredential(ctx context.Context, logger logr.Logger, cfg Config) (azcore.TokenCredential, error) {
	if cfg.AzureKVManagedClientID != "" || cfg.AzureKVManagedResourceID != "" {
		opts := azidentity.ManagedIdentityCredentialOptions{}
		if cfg.AzureKVManagedClientID != "" {
			opts.ID = azidentity.ClientID(cfg.AzureKVManagedClientID)
		} else if cfg.AzureKVManagedResourceID != "" {
			opts.ID = azidentity.ResourceID(cfg.AzureKVManagedResourceID)
		}

		managed, err := azidentity.NewManagedIdentityCredential(&opts)
		if err == nil {
			logger.Info("Azure Managed Identity will be used as authentication method")
			return managed, err
		}
	}

	spSecret, err := azidentity.NewClientSecretCredential(cfg.AzureKVTenantID, cfg.AzureKVClientID, cfg.AzureKVClientSecret, nil)
	if err == nil {
		logger.Info("Azure Service Principal will be used as authentication method")
		return spSecret, err
	}

	return nil, goerrors.New("Unable to authenticate to Azure API using any method")
}

func azureKeyVaultClient(ctx context.Context, l logr.Logger, cfg Config) (*azureKVClient, error) {
	logger := l.WithName("azure-kv").WithValues(
		"azure_kv_name", cfg.AzureKVName,
		"azure_kv_tenant", cfg.AzureKVTenantID)

	cred, err := getAzureCredential(ctx, logger, cfg)
	if err != nil {
		logger.Error(err, "Error occured while authenticating to Azure")
		return nil, err
	}
	akvMetrics = newAzureKVMetrics(cfg.AzureKVName, cfg.AzureKVTenantID)
	vaultEndpoint := fmt.Sprintf("https://%s.%s", cfg.AzureKVName, azureKVEndpoint)
	akvClient, err := azsecrets.NewClient(vaultEndpoint, cred, nil)

	if err != nil {
		logger.Error(err, "Error occured while creating Azure KV client")
		akvMetrics.updateLoginErrorsTotalMetric()
		return nil, err
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
		var responseError *azcore.ResponseError
		if goerrors.As(err, &responseError) {
			if responseError.StatusCode == 404 {
				errorType = errors.BackendSecretNotFoundErrorType
			}
			if responseError.StatusCode == 403 {
				errorType = errors.BackendSecretForbiddenErrorType
			}
		}
		akvMetrics.updateSecretReadErrorsTotalMetric(path, errorType)
		return data, err
	}

	data = *result.Value
	return data, err
}
