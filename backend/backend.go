package backend

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/tuenti/secrets-manager/errors"
)

const vaultBackendName = "vault"

var supportedBackends map[string]bool

func init() {
	supportedBackends = map[string]bool{vaultBackendName: true}
}

// Config type represent backend config, and should include all backends config
type Config struct {
	BackendTimeout          time.Duration
	VaultURL                string
	VaultRoleID             string
	VaultSecretID           string
	VaultMaxTokenTTL        int64
	VaultTokenPollingPeriod time.Duration
	VaultRenewTTLIncrement  int
	VaultEngine             string
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
	}
	return &client, err
}
