package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/tuenti/secrets-manager/errors"
)

var (
	logger  *log.Logger
	metrics *vaultMetrics
)

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
}

func (c *client) vaultLogin() error {
	appRole := map[string]interface{}{
		"role_id":   c.roleID,
		"secret_id": c.secretID,
	}
	resp, err := c.logical.Write("auth/approle/login", appRole)
	if err != nil {
		logger.Errorf("unable to login to Vault: %v", err)
		return err
	}
	c.vclient.SetToken(resp.Auth.ClientToken)
	return nil
}

func vaultClient(l *log.Logger, cfg Config) (*client, error) {
	if l != nil {
		logger = l
	} else {
		logger = log.New()
	}

	httpClient := new(http.Client)
	httpClient.Timeout = cfg.BackendTimeout

	vclient, err := api.NewClient(&api.Config{Address: cfg.VaultURL, HttpClient: httpClient})

	if err != nil {
		logger.Debugf("unable to build vault client: %v", err)
		return nil, err
	}

	logical := vclient.Logical()

	engine, err := newEngine(cfg.VaultEngine)
	if err != nil {
		logger.Debugf("unable to use engine %s: %v", cfg.VaultEngine, err)
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
		logger.Debugf("unable to login with provided credentials: %v", err)
		return nil, err
	}

	sys := vclient.Sys()
	health, err := sys.Health()

	if err != nil {
		logger.Debugf("could not contact Vault at %s: %v ", cfg.VaultURL, err)
		return nil, err
	}

	logger.Infof("successfully logged into Vault cluster %s", health.ClusterName)

	metrics = newVaultMetrics(cfg.VaultURL, health.Version, cfg.VaultEngine, health.ClusterID, health.ClusterName)

	metrics.updateVaultMaxTokenTTLMetric(cfg.VaultMaxTokenTTL)

	return &client, err
}

func (c *client) getToken() (*api.Secret, error) {
	auth := c.vclient.Auth()
	lookup, err := auth.Token().LookupSelf()
	if err != nil {
		logger.Errorf("error checking token with lookup self api: %v", err)
		metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultLookupSelfOperationName, errors.UnknownErrorType)
		return nil, err
	}
	return lookup, nil
}

func (c *client) getTokenTTL(token *api.Secret) (int64, error) {
	var ttl int64
	ttl, err := token.Data["ttl"].(json.Number).Int64()
	if err != nil {
		logger.Errorf("couldn't decode ttl from token: %v", err)
		return -1, err
	}
	metrics.updateVaultTokenTTLMetric(ttl)
	return ttl, nil
}

func (c *client) renewToken(token *api.Secret) error {
	isRenewable, err := token.TokenIsRenewable()
	if err != nil {
		logger.Errorf("could not check token renewability: %v", err)
		metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultIsRenewableOperationName, errors.UnknownErrorType)
		return err
	}
	if !isRenewable {
		metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultIsRenewableOperationName, errors.VaultTokenNotRenewableErrorType)
		err = &errors.VaultTokenNotRenewableError{ErrType: errors.VaultTokenNotRenewableErrorType}
		return err
	}
	auth := c.vclient.Auth()
	if _, err = auth.Token().RenewSelf(c.renewTTLIncrement); err != nil {
		log.Errorf("failed to renew token: %v", err)
		metrics.updateVaultTokenRenewalErrorsTotalMetric(vaultRenewSelfOperationName, errors.UnknownErrorType)
		return err
	}
	return nil
}

func (c *client) renewalLoop() {
	token, err := c.getToken()
	if err != nil {
		logger.Errorf("failed to fetch token: %v", err)
		logger.Warnf("trying to login again")
		if err = c.vaultLogin(); err != nil {
			metrics.updateVaultLoginErrorsTotalMetric()
		}
		return
	}
	ttl, err := c.getTokenTTL(token)
	if err != nil {
		logger.Errorf("failed to read token TTL: %v", err)
		return
	} else if ttl < c.maxTokenTTL {
		logger.Warnf("token is really close to expire, current ttl: %d", ttl)
		err := c.renewToken(token)
		if err != nil {
			logger.Errorf("could not renew token: %v", err)
		} else {
			logger.Infoln("token renewed successfully!")
		}
	} else {
		return
	}
}

func (c *client) startTokenRenewer(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			select {
			case <-time.After(c.tokenPollingPeriod):
				c.renewalLoop()
				break
			case <-ctx.Done():
				logger.Infoln("gracefully shutting down token renewal go routine")
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
		metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.UnknownErrorType)
		return data, err
	}

	if secret != nil {
		secretData := c.engine.getData(secret)
		warnings := secret.Warnings
		if secretData != nil {
			if secretData[key] != nil {
				data = secretData[key].(string)
			} else {
				metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
				err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
			}
		} else {
			for _, w := range warnings {
				logger.Warningln(w)
			}
			metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
			err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
		}
	} else {
		metrics.updateVaultSecretReadErrorsTotalMetric(path, key, errors.BackendSecretNotFoundErrorType)
		err = &errors.BackendSecretNotFoundError{ErrType: errors.BackendSecretNotFoundErrorType, Path: path, Key: key}
	}
	return data, err
}
