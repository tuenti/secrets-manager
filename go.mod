module github.com/tuenti/secrets-manager

go 1.16

require (
	github.com/Azure/azure-sdk-for-go v63.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.1
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.1
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/go-logr/logr v0.3.0
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/vault/api v1.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
