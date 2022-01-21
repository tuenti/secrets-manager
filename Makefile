DOCKER_REGISTRY = "registry.hub.docker.com"
ORGANIZATION = "tuentitech"
BINARY_NAME=secrets-manager
VERSION=$(shell deploy/version/get.sh)
GO111MODULE=on
# Image URL to use all building/pushing image targets
IMAGE = ${DOCKER_REGISTRY}/${ORGANIZATION}/${BINARY_NAME}:${VERSION}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

BUILD_FLAGS=-ldflags "-X main.version=${VERSION}"

all: manager

# Run tests
test: generate fmt vet manifests
	go test -v ./backend/... ./errors/... ./controllers/... -coverprofile cover.out

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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Run tests in docker
docker-test:
	docker-compose run tests

# Build release docker image
docker-build: docker-test
	docker build . \
		--file ./deploy/Dockerfile \
		--target release \
		--build-arg SECRETS_MANAGER_VERSION=${VERSION} \
		--tag ${IMAGE}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMAGE}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMAGE}

update-major-version:
	deploy/version/update.sh --minor

update-minor-version:
	deploy/version/update.sh --minor

update-patch-version:
	deploy/version/update.sh --patch
