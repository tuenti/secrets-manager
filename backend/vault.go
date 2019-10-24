package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	"github.com/tuenti/secrets-manager/errors"
)

var vMetrics *vaultMetrics

const defaultSecretKey = "data"

type client struct {
	vclient            *api.Client
	logical            *api.Logical
	roleID             string
	secretID           string
	maxTokenTTL        int64
	tokenPollingPeriod time.Duration
	renewTTLIncrement  int
	engine             engine
	logger             logr.Logger
}

func (c *client) vaultLogin() error {
	appRole := map[string]interface{}{
		"role_id":   c.roleID,
		"secret_id": c.secretID,
	}
	resp, err := c.logical.Write("auth/approle/login", appRole)
	if err != nil {
		return err
	}
	c.vclient.SetToken(resp.Auth.ClientToken)
	return nil
}

func vaultClient(l logr.Logger, cfg Config) (*client, error) {
	logger := l.WithName("vault").WithValues(
		"vault_url", cfg.VaultURL,
		"vault_engine", cfg.VaultEngine)

	httpClient := new(http.Client)
	httpClient.Timeout = cfg.BackendTimeout

	vclient, err := api.NewClient(&api.Config{Address: cfg.VaultURL, HttpClient: httpClient})

	if err != nil {
		logger.Error(err, "unable to create vault api client")
		return nil, err
	}

	logical := vclient.Logical()

	engine, err := newEngine(cfg.VaultEngine)
	if err != nil {
		logger.Error(err, "unable to setup vault engine")
		return nil, err
	}

	client := client{
		vclient:            vclient,
		logical:            logical,
		roleID:             cfg.VaultRoleID,
		secretID:           cfg.VaultSecretID,
		maxTokenTTL:        cfg.VaultMaxTokenTTL,
		tokenPollingPeriod: cfg.VaultTokenPollingPeriod,
		renewTTLIncrement:  cfg.VaultRenewTTLIncrement,
		engine:             engine,
	}

	err = client.vaultLogin()
	if err != nil {
		logger.Error(err, "unable to login to vault with provided credentials")
		return nil, err
	}

	sys := vclient.Sys()
	health, err := sys.Health()

	if err != nil {
		logger.Error(err, "could not get health information about vault cluster")
		return nil, err
	}

	logger = logger.WithValues(
		"vault_cluster_name", health.ClusterName,
		"vault_cluster_id", health.ClusterID,
		"vault_version", health.Version,
		"vault_sealed", strconv.FormatBool(health.Sealed),
		"vault_server_time_utc", health.ServerTimeUTC,
	)

	logger.Info("successfully logged into vault cluster")

	client.logger = logger

	vMetrics = newVaultMetrics(cfg.VaultURL, health.Version, cfg.VaultEngine, health.ClusterID, health.ClusterName)

	vMetrics.updateVaultMaxTokenTTLMetric(cfg.VaultMaxTokenTTL)

	return &client, err
}

func (c *client) getToken() (*api.Secret, error) {
	auth := c.vclient.Auth()
	lookup, err := auth.Token().LookupSelf()
	if err != nil {
		vMetrics.updateVaultTokenRenewalErrorsTotalMetric(vaultLookupSelfOperationName, errors.UnknownErrorType)
		return nil, err
	}
	return lookup, nil
}

func (c *client) getTokenTTL(token *api.Secret) (int64, error) {
	var ttl int64
	ttl, err := token.Data["ttl"].(json.Number).Int64()
	if err != nil {
		return -1, err
	}
	vMetrics.updateVaultTokenTTLMetric(ttl)
	return ttl, nil
}

func (c *client) renewToken(token *api.Secret) error {
	isRenewable, err := token.TokenIsRenewable()
	if err != nil {
		vMetrics.updateVaultTokenRenewalErrorsTotalMetric(vaultIsRenewableOperationName, errors.UnknownErrorType)
		return err
	}
	if !isRenewable {
		vMetrics.updateVaultTokenRenewalErrorsTotalMetric(vaultIsRenewableOperationName, errors.VaultTokenNotRenewableErrorType)
		err = &errors.VaultTokenNotRenewableError{ErrType: errors.VaultTokenNotRenewableErrorType}
		return err
	}
	auth := c.vclient.Auth()
	if _, err = auth.Token().RenewSelf(c.renewTTLIncrement); err != nil {
		vMetrics.updateVaultTokenRenewalErrorsTotalMetric(vaultRenewSelfOperationName, errors.UnknownErrorType)
		return err
	}
	return nil
}

func (c *client) renewalLoop() {
	token, err := c.getToken()
	if err != nil {
		c.logger.Error(err, "unable to get vault token")
		c.logger.Info("trying to login to vault again")
		if err = c.vaultLogin(); err != nil {
			vMetrics.updateVaultLoginErrorsTotalMetric()
			c.logger.Info("login error, vault token not obtained")
		} else {
			c.logger.Info("login successful, got a new vault token")
		}
		return
	}

	ttl, err := c.getTokenTTL(token)
	if err != nil {
		c.logger.Error(err, "failed to read vault token TTL")
	} else if ttl < c.maxTokenTTL {
		c.logger.Info("vault token is really close to expire", "vault_token_ttl", ttl)
		err := c.renewToken(token)
		if err != nil {
			c.logger.Error(err, "failed to renew vault token")
		} else {
			c.logger.Info("vault token renewed successfully!")
		}
	}
	return
}

func (c *client) startTokenRenewer(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-time.After(c.tokenPollingPeriod):
				c.renewalLoop()
				break
			case <-ctx.Done():
				c.logger.Info("gracefully shutting down token renewal go routine")
				return
			}
		}
	}(ctx)
}

func (c *client) ReadSecret(path string, key string) (string, error) {
	data := ""
	if key == "" {
		key = defaultSecretKey
	}

	logical := c.logical
	secret, err := logical.Read(path)
	if err != nil {
		vMetrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.UnknownErrorType)
		return data, err
	}

	if secret != nil {
		secretData := c.engine.getData(secret)
		warnings := secret.Warnings
		if secretData != nil {
			if secretData[key] != nil {
				data = secretData[key].(string)
			} else {
				vMetrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
				err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
			}
		} else {
			for _, w := range warnings {
				c.logger.Info("secret contains warnings", "vault_secret_warning", w)
			}
			vMetrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
			err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
		}
	} else {
		vMetrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
		err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
	}
	return data, err
}
