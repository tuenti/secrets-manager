
# Binary names
BUILD_FOLDER=build
BINARY_NAME=secrets-manager
SECRETS_MANAGER_VERSION=v1.0.0-snapshot
DOCKER_REGISTRY ?= "registry.hub.docker.com"
DOCKER_IMAGE_NAME=${BINARY_NAME}
DOCKER_IMAGE=${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${SECRETS_MANAGER_VERSION}
BUILD_FLAGS=-ldflags "-X main.version=${SECRETS_MANAGER_VERSION}"

pkgs   = $(shell go list ./... | grep -v /vendor/)

.PHONY: init
init:
	scripts/setup-dev-env.sh
	glide install

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${BUILD_FOLDER}/${BINARY_NAME} -v

.PHONY: docker-build
docker-build:
	docker build --build-arg SECRETS_MANAGER_VERSION=${SECRETS_MANAGER_VERSION} -t ${DOCKER_IMAGE} .

.PHONY: docker-push
docker-push: docker-build
	docker login ${DOCKER_REGISTRY}
	docker push ${DOCKER_IMAGE}

.PHONY: test
test: mocks
	mkdir -p ${BUILD_FOLDER}
	go test -coverprofile=${BUILD_FOLDER}/coverage.txt ./... 
	go tool cover -html=${BUILD_FOLDER}/coverage.txt -o ${BUILD_FOLDER}/coverage.html
	go tool cover -func=${BUILD_FOLDER}/coverage.txt

.PHONY: style
style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: format
format:
	@echo ">> formatting code"
	@go fmt $(pkgs)

.PHONY: vet
vet:
	@echo ">> vetting code"
	@go vet $(pkgs)

mocks:
	mockgen -package mocks \
		-destination mocks/kubernetes.go \
		-source kubernetes/kubernetes.go \
		-mock_names Client=MockKubernetesClient
