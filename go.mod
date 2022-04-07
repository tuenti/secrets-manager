module github.com/tuenti/secrets-manager

go 1.16

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.22.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.13.2
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v0.6.0
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/vault/api v1.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
