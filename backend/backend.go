package backend

import (
	"context"
	"time"

	"github.com/tuenti/secrets-manager/errors"

	log "github.com/sirupsen/logrus"
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
	VaultToken              string
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
func NewBackendClient(ctx context.Context, backend string, logger *log.Logger, cfg Config) (*Client, error) {
	var err error
	var client Client

	if !supportedBackends[backend] {
		err = &errors.BackendNotImplementedError{ErrType: errors.BackendNotImplementedErrorType, Backend: backend}
		return nil, err
	}
	switch backend {
	case vaultBackendName:
		client, err = vaultClient(ctx, logger, cfg)
	}
	return &client, err
}
