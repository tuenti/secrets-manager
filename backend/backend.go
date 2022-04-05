package backend

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/tuenti/secrets-manager/errors"
)

const (
	vaultBackendName   = "vault"
	azureKVBackendName = "azure-kv"
)

var supportedBackends map[string]bool

func init() {
	supportedBackends = map[string]bool{
		vaultBackendName:   true,
		azureKVBackendName: true,
	}
}

// Config type represent backend config, and should include all backends config
type Config struct {
	BackendTimeout           time.Duration
	VaultURL                 string
	VaultAuthMethod          string
	VaultRoleID              string
	VaultSecretID            string
	VaultKubernetesRole      string
	VaultMaxTokenTTL         int64
	VaultTokenPollingPeriod  time.Duration
	VaultRenewTTLIncrement   int
	VaultEngine              string
	VaultApprolePath         string
	VaultKubernetesPath      string
	AzureKVName              string
	AzureKVTenantID          string
	AzureKVClientID          string
	AzureKVClientSecret      string
	AzureKVManagedClientID   string
	AzureKVManagedResourceID string
}

// Client interface represent a backend client interface that should be implemented
type Client interface {
	ReadSecret(path string, key string) (string, error)
}

// NewBackendClient returns and implementation of Client interface, given the selected backend
func NewBackendClient(ctx context.Context, backend string, logger logr.Logger, cfg Config) (*Client, error) {
	var err error
	var client Client

	if !supportedBackends[backend] {
		err = &errors.BackendNotImplementedError{ErrType: errors.BackendNotImplementedErrorType, Backend: backend}
		return nil, err
	}
	switch backend {
	case vaultBackendName:
		vclient, verr := vaultClient(logger, cfg)
		if verr != nil {
			return nil, verr
		}
		vclient.startTokenRenewer(ctx)
		client = vclient
		err = verr
	case azureKVBackendName:
		akvclient, akverr := azureKeyVaultClient(ctx, logger, cfg)
		if akverr != nil {
			return nil, akverr
		}
		client = akvclient
		err = akverr
	}
	return &client, err
}
