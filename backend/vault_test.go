package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuenti/secrets-manager/errors"

	"github.com/gorilla/mux"
)

const (
	vaultAPIVersion       = "v1"
	selectedBackend       = "vault"
	fakeToken             = "fake-token"
	defaultTokenTTL       = 40
	defaultTokenRenewable = true
)

type testConfig struct {
	tokenTTL       int
	tokenRenewable bool
}

var (
	vaultCfg Config
	ctx      context.Context
	server   *httptest.Server
	mutex    sync.Mutex
	testCfg  *testConfig
)

func v1SysHealth(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	jsonData := `
	{
		"initialized": true,
		"sealed": false,
		"standby": false,
		"performance_standby": false,
		"replication_performance_mode": "disabled",
		"replication_dr_mode": "disabled",
		"server_time_utc": 1537804485,
		"version": "0.11.1",
		"cluster_name": "vault-cluster-mock",
		"cluster_id": "mock-0"
	}`

	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func v1AuthTokenLookupSelf(w http.ResponseWriter, r *http.Request) {
	var response interface{}

	jsonData := fmt.Sprintf(`
	{
		"request_id": "8d70f864-5f77-44fe-0940-df085376101f",
		"lease_id": "",
		"renewable": false,
		"lease_duration": 0,
		"data": {
			"accessor": "d2d7308c-b9f2-3399-4202-11d670b8c053",
			"creation_time": 1537810558,
			"creation_ttl": 60,
			"display_name": "token",
			"entity_id": "",
			"expire_time": "2018-09-24T17:36:58.797772932Z",
			"explicit_max_ttl": 0,
			"id": "31a5ea4e-907d-c1b9-1dfc-6b88526be248",
			"issue_time": "2018-09-24T17:35:58.79776585Z",
			"meta": null,
			"num_uses": 0,
			"orphan": false,
			"path": "auth/token/create",
			"policies": [
				"fake-policy"
			],
			"renewable": %t,
			"ttl": %d
		},
		"wrap_info": null,
		"warnings": null,
		"auth": null
	}`, testCfg.tokenRenewable, testCfg.tokenTTL)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func v1AuthTokenRenewSelf(w http.ResponseWriter, r *http.Request) {
	var response interface{}

	jsonData := fmt.Sprintf(`{
		"request_id": "d8ae3e67-91a0-2f7a-528b-522048f9dad3",
		"lease_id": "",
		"renewable": false,
		"lease_duration": 0,
		"data": null,
		"wrap_info": null,
		"warnings": null,
		"auth": {
			"client_token": "%s",
			"accessor": "dc6aa861-3020-322c-8df5-4b08afa43a34",
			"policies": [
				"fake-policy"
			],
			"token_policies": [
				"fake-policy"
			],
			"metadata": null,
			"lease_duration": 1000,
			"renewable": true,
			"entity_id": ""
		}
	}`, fakeToken)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func v1SecretTestKv2(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	jsonData := `
	{
		"request_id": "a21f835e-7e72-dd43-d5a1-80fea23c0649",
		"lease_id": "",
		"renewable": false,
		"lease_duration": 0,
		"data": {
			"data": {
				"foo": "bar"
			},
			"metadata": {
				"created_time": "2018-09-25T08:35:15.504392904Z",
				"deletion_time": "",
				"destroyed": false,
				"version": 1
			}
		},
		"wrap_info": null,
		"warnings": null,
		"auth": null
	}`
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func v1SecretTestKv1(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	jsonData := `
	{
		"request_id": "a21f835e-7e72-dd43-d5a1-80fea23c0649",
		"lease_id": "",
		"renewable": false,
		"lease_duration": 0,
		"data": {
			"foo": "bar",
		},
		"wrap_info": null,
		"warnings": null,
		"auth": null
	}`
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func TestVaultClient(t *testing.T) {
	client, err := vaultClient(ctx, nil, vaultCfg)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestTokenNotExpired(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	mutex.Lock()
	defer mutex.Unlock()
	testCfg.tokenRenewable = true
	testCfg.tokenTTL = 600
	client.maxTokenTTL = 60
	assert.False(t, client.isTokenExpired())
}

func TestTokenExpired(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	mutex.Lock()
	defer mutex.Unlock()
	testCfg.tokenRenewable = true
	testCfg.tokenTTL = 10
	client.maxTokenTTL = 50
	assert.True(t, client.isTokenExpired())
}

func TestTokenNotRenewableExpired(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	mutex.Lock()
	defer mutex.Unlock()
	testCfg.tokenRenewable = false
	assert.False(t, client.isTokenExpired())
}

func TestRenewToken(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	err := client.renewToken()
	assert.Nil(t, err)
}

func TestReadSecretKv2(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	secretValue, err := client.ReadSecret("/secret/data/test", "foo")
	assert.Nil(t, err)
	assert.Equal(t, "bar", secretValue)
}

func TestReadSecretKv1(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()
	vaultCfg.VaultEngine = "kv1"
	client, _ := vaultClient(ctx, nil, vaultCfg)
	secretValue, err := client.ReadSecret("/secret/test", "foo")
	assert.Nil(t, err)
	assert.Equal(t, "bar", secretValue)
}

func TestSecretNotFound(t *testing.T) {
	client, _ := vaultClient(ctx, nil, vaultCfg)
	secretValue, err := client.ReadSecret("/secret/data/test", "foo2")
	assert.Empty(t, secretValue)
	assert.EqualError(t, err, fmt.Sprintf("[%s] secret key %s not found at %s", errors.BackendSecretNotFoundErrorType, "foo2", "/secret/data/test"))
}

func TestMain(m *testing.M) {
	r := mux.NewRouter()
	v1SysHandler := r.PathPrefix(fmt.Sprintf("/%s/sys", vaultAPIVersion)).Subrouter()
	v1AuthHandler := r.PathPrefix(fmt.Sprintf("/%s/auth", vaultAPIVersion)).Subrouter()
	v1SecretHandler := r.PathPrefix(fmt.Sprintf("/%s/secret", vaultAPIVersion)).Subrouter()

	v1SysHandler.HandleFunc("/health", v1SysHealth).Methods("GET")
	v1AuthHandler.HandleFunc("/token/lookup-self", v1AuthTokenLookupSelf).Methods("GET")
	v1AuthHandler.HandleFunc("/token/renew-self", v1AuthTokenRenewSelf).Methods("PUT")
	v1SecretHandler.HandleFunc("/data/test", v1SecretTestKv2).Methods("GET")
	v1SecretHandler.HandleFunc("/test", v1SecretTestKv1).Methods("GET")

	server = httptest.NewServer(r)
	defer server.Close()

	vaultCfg = Config{
		VaultURL:                string(server.URL),
		VaultToken:              fakeToken,
		VaultTokenPollingPeriod: 1,
	}

	testCfg = &testConfig{
		tokenRenewable: defaultTokenRenewable,
		tokenTTL:       defaultTokenTTL,
	}

	_ctx, cancel := context.WithCancel(context.Background())
	ctx = _ctx
	defer cancel()

	m.Run()
}
