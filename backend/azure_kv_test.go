package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

const (
	fakeKeyVaultName   = "azure-keyvault-fake-name"
	fakeKeyVaultTenant = "01234567-0123-0123-0123-0123456789ab"
	fakeKeyVaultSecret = "fake-secret"
)

var akvSecrets = map[string]struct {
	value  string
	access bool
}{
	fakeKeyVaultSecret: {value: "some-fake-value", access: true},
	"exists":           {value: "yes", access: true},
	"internal-error":   {value: "\"bad-scaped", access: true},
	"forbidden":        {value: "yes", access: false},
}

func akvGetSecret(w http.ResponseWriter, r *http.Request) {
	// Info about what keyvault responses should look like extracted from
	// https://github.com/Azure/azure-sdk-for-go/blob/c73b114ded83c0a9c2685336b8b90836c1530cb3/sdk/keyvault/azsecrets/testdata/recordings/TestSetGetSecret.json
	vars := mux.Vars(r)
	jsonData := "{}"
	if v, ok := akvSecrets[vars["secretName"]]; ok {
		if v.access {
			jsonData = fmt.Sprintf(`
			{
				"value": "%s",
				"id": "https://%s.vault.azure.net/secrets/%s/3f3b11064811494a8a8b27edf4f0985b",
        "attributes": {
          "enabled": true,
          "created": 1643130727,
          "updated": 1643130727,
          "recoveryLevel": "CustomizedRecoverable\u002BPurgeable",
          "recoverableDays": 7
        }
			}`, v.value, fakeKeyVaultName, vars["secretName"])
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	} else {
		jsonData = fmt.Sprintf(`
		"error": {
          "code": "SecretNotFound",
          "message": "Secret not found: %s"
        }
		`, vars["secretName"])
		w.WriteHeader(http.StatusNotFound)
	}
	var response interface{}

	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Printf("unable to unmarshal json %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-ms-keyvault-network-info", "conn_type=Ipv4;addr=72.49.29.93;act_addr_fam=InterNetwork;")
	w.Header().Set("x-ms-keyvault-region", "westus2")
	w.Header().Set("x-ms-keyvault-service-version", "1.9.264.2")
	w.Header().Set("x-ms-request-id", "868ba1d2-efe7-4930-b3ad-d010cf499778")
	w.Header().Set("X-Powered-By", "ASP.NET")
	// Trick Azure client to make it think everything is legit
	w.Header().Set(
		"WWW-Authenticate",
		"Bearer authorization=\u0022https://login.windows.net/72f988bf-86f1-41af-91ab-2d7cd011db47\u0022, resource=\u0022https://vault.azure.net\u0022",
	)
	json.NewEncoder(w).Encode(response)
}

// Copied from https://github.com/Azure/azure-sdk-for-go/blob/35fb64f82ef3b3308f55b1da37c1fec36bdd4166/sdk/keyvault/azsecrets/utils_test.go
type FakeCredential struct {
	accountName string
	accountKey  string
}

func NewFakeCredential(accountName, accountKey string) *FakeCredential {
	return &FakeCredential{
		accountName: accountName,
		accountKey:  accountKey,
	}
}

func (f *FakeCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (*azcore.AccessToken, error) {
	return &azcore.AccessToken{
		Token:     "faketoken",
		ExpiresOn: time.Date(2040, time.January, 1, 1, 1, 1, 1, time.UTC),
	}, nil
}

func TestGetAzureCredential(t *testing.T) {
	cases := []struct {
		cfg Config
		err bool
		typ azcore.TokenCredential
		msg string
	}{
		{
			Config{},
			true,
			nil,
			"Empty config should not be able to generate any client",
		},
		{
			Config{AzureKVManagedClientID: "fake-client-id"},
			false,
			new(azidentity.ManagedIdentityCredential),
			"Managed identity client should be generated using managed client ID",
		},
		{
			Config{AzureKVManagedResourceID: "fake-resource-id"},
			false,
			new(azidentity.ManagedIdentityCredential),
			"Managed identity client should be generated using managed resource ID",
		},
		{
			Config{
				AzureKVTenantID: "fake-tenant-id",
			},
			true,
			nil,
			"Incomplete config should not generate any client (tenant)",
		},
		{
			Config{
				AzureKVClientID: "fake-client-id",
			},
			true,
			nil,
			"Incomplete config should not generate any client (clientID)",
		},
		{
			Config{
				AzureKVClientSecret: "fake-client-secret",
			},
			true,
			nil,
			"Incomplete config should not generate any client (clientID)",
		},
		{
			Config{
				AzureKVTenantID: "fake-tenant-id",
				AzureKVClientID: "fake-client-id",
			},
			true,
			nil,
			"Incomplete config should not generate any client (tenant, clientID)",
		},
		{
			Config{
				AzureKVClientID:     "fake-client-id",
				AzureKVClientSecret: "fake-client-secret",
			},
			true,
			nil,
			"Incomplete config should not generate any client (clientID, clientSecret)",
		},
		{
			Config{
				AzureKVTenantID:     "fake-tenant-id",
				AzureKVClientID:     "fake-client-id",
				AzureKVClientSecret: "fake-client-secret",
			},
			false,
			new(azidentity.ClientSecretCredential),
			"ClientSecretCredential should be generated with TenantID, ClientID and ClientSecret",
		},
	}
	for _, c := range cases {
		client, err := getAzureCredential(context.TODO(), logger, c.cfg)
		if c.err {
			assert.NotNilf(t, err, c.msg)
		} else {
			assert.Nilf(t, err, c.msg)
		}
		if c.typ == nil {
			assert.Nilf(t, client, c.msg)
		} else {
			assert.IsTypef(t, c.typ, client, c.msg)
		}
	}
}

func TestAzureKeyVaultClient(t *testing.T) {
	cfg := Config{}
	client, err := azureKeyVaultClient(context.TODO(), logger, cfg)
	assert.NotNilf(t, err, "Empty config should generate an error")
	assert.Nilf(t, client, "Empty config should not generate any client")

	// Managed Identity auth is performed at client call, so the client generated is "valid"
	cfg = Config{
		AzureKVTenantID: fakeKeyVaultTenant,
		AzureKVClientID: "fake-client-id",
	}
	client, err = azureKeyVaultClient(context.TODO(), logger, cfg)
	assert.NotNilf(t, err, "Invalid Service Principal Authentication should generate an error")
	assert.Nilf(t, client, "Invalid Service Principal Authentication should not generate any client")

	// Authentication is performed at client call, so the client generated is "valid"
	// This happens for both service principal and managed identity
	cfg = Config{
		AzureKVTenantID:     fakeKeyVaultTenant,
		AzureKVClientID:     "fake-client-id",
		AzureKVClientSecret: "fake-client-secret",
	}
	client, err = azureKeyVaultClient(context.TODO(), logger, cfg)
	assert.Nilf(t, err, "Service Principal Authentication should not generate error")
	assert.NotNilf(t, client, "Service Principal Authentication should generate a client")

	cfg = Config{AzureKVManagedClientID: "fake-client-id"}
	client, err = azureKeyVaultClient(context.TODO(), logger, cfg)
	assert.Nilf(t, err, "Managed Identity Authentication should not generate error")
	assert.NotNilf(t, client, "Managed Identity Authentication should generate a client")
}

func TestAzureKVClientReadSecret(t *testing.T) {
	akvMetrics = newAzureKVMetrics(fakeKeyVaultName, fakeKeyVaultTenant)
	azClient, _ := azsecrets.NewClient(
		testingCfg.VaultURL, // Is a mock server, valid for both cases
		NewFakeCredential("fake", "fake"),
		nil,
	)
	client := azureKVClient{
		client:       azClient,
		keyvaultName: "fakekvurl",
		context:      context.TODO(),
		logger:       logger,
	}

	value, err := client.ReadSecret("exists", "")
	assert.Nil(t, err)
	assert.Equal(t, akvSecrets["exists"].value, value)

	value, err = client.ReadSecret(fakeKeyVaultSecret, "")
	assert.Nil(t, err)
	assert.Equal(t, akvSecrets[fakeKeyVaultSecret].value, value)

	value, err = client.ReadSecret("forbidden", "")
	assert.NotNil(t, err)
	assert.Equal(t, "", value)
	assert.IsType(t, new(azcore.ResponseError), err)

	value, err = client.ReadSecret("not-found", "")
	assert.NotNil(t, err)
	assert.Equal(t, "", value)
	assert.IsType(t, new(azcore.ResponseError), err)

	value, err = client.ReadSecret("internal-error", "")
	assert.NotNil(t, err)
	assert.Equal(t, "", value)
}
