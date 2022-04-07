package backend

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testingCfg Config
	server     *httptest.Server
	mutex      sync.Mutex
	logger     logr.Logger
)

func TestNotImplementedBackend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := Config{}
	backend := "foo"
	_, err := NewBackendClient(ctx, backend, nil, cfg)
	assert.EqualError(t, err, fmt.Sprintf("[%s] backend %s not supported", errors.BackendNotImplementedErrorType, backend))
}

func TestMain(m *testing.M) {
	r := mux.NewRouter()
	v1SysHandler := r.PathPrefix(fmt.Sprintf("/%s/sys", vaultAPIVersion)).Subrouter()
	v1AuthHandler := r.PathPrefix(fmt.Sprintf("/%s/auth", vaultAPIVersion)).Subrouter()
	v1SecretHandler := r.PathPrefix(fmt.Sprintf("/%s/secret", vaultAPIVersion)).Subrouter()
	akvSecretsHandler := r.PathPrefix("/secrets").Subrouter()

	v1SysHandler.HandleFunc("/health", v1SysHealth).Methods("GET")
	v1AuthHandler.HandleFunc("/token/lookup-self", v1AuthTokenLookupSelf).Methods("GET")
	v1AuthHandler.HandleFunc("/token/renew-self", v1AuthTokenRenewSelf).Methods("PUT")
	v1AuthHandler.HandleFunc("/approle/login", v1AuthAppRoleLogin).Methods("PUT")
	v1AuthHandler.HandleFunc("/kubernetes/login", v1AuthKubernetesLogin).Methods("PUT")
	v1SecretHandler.HandleFunc("/data/test", v1SecretTestKv2).Methods("GET")
	v1SecretHandler.HandleFunc("/test", v1SecretTestKv1).Methods("GET")

	akvSecretsHandler.PathPrefix("/{secretName}").HandlerFunc(akvGetSecret).Methods("GET")

	server = httptest.NewServer(r)
	defer server.Close()

	testingCfg = Config{
		VaultURL:                string(server.URL),
		VaultRoleID:             vaultFakeRoleID,
		VaultSecretID:           vaultFakeSecretID,
		VaultTokenPollingPeriod: 1,
		VaultEngine:             "kv2",
		VaultApprolePath:        vaultAppRolePath,
	}

	vaultTestCfg = &testConfig{
		tokenRenewable:  defaultTokenRenewable,
		tokenTTL:        defaultTokenTTL,
		tokenRevoked:    defaultRevokedToken,
		invalidRoleID:   defaultInvalidAppRole,
		invalidSecretID: defaultInvalidAppRole,
	}

	logger = zap.New(zap.UseDevMode(true))

	os.Exit(m.Run())
}
