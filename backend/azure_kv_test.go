package backend

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
)

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
