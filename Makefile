DOCKER_REGISTRY ?= "registry.hub.docker.com"
BINARY_NAME=secrets-manager
SECRETS_MANAGER_VERSION=v1.0.0
GO111MODULE=on
# Image URL to use all building/pushing image targets
IMG = ${DOCKER_REGISTRY}/${BINARY_NAME}:${SECRETS_MANAGER_VERSION}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

BUILD_FLAGS=-ldflags "-X main.version=${SECRETS_MANAGER_VERSION}"

all: manager

# Run tests
test: generate fmt vet manifests
	go test -v ./backend... ./errors/... ./controllers/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build ${BUILD_FLAGS} -o bin/${BINARY_NAME} main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# Build the docker image
docker-build: test
	docker build --build-arg SECRETS_MANAGER_VERSION=${SECRETS_MANAGER_VERSION} -t ${IMG} .
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0-beta.2
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
