# Build the manager binary
FROM golang:1.12.5 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY backend/ backend/
COPY errors/ errors/
COPY hack/ hack/
ARG SECRETS_MANAGER_VERSION

# Build
RUN go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0-beta.2
RUN /go/bin/controller-gen object:headerFile=./hack/boilerplate.go.txt paths=./api/...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-X main.version=${SECRETS_MANAGER_VERSION}" -a -o secrets-manager main.go


#
# Prod image
#

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest as release
WORKDIR /
COPY --from=builder /workspace/secrets-manager .
ENTRYPOINT ["/secrets-manager"]
USER nobody


#
# Dev image
#

FROM builder as dev

# kubebuilder needed to run tests in development environment
RUN curl -L -O https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.0.0-alpha.4/kubebuilder_2.0.0-alpha.4_linux_amd64.tar.gz
RUN tar -zxvf kubebuilder_2.0.0-alpha.4_linux_amd64.tar.gz
RUN mv kubebuilder_2.0.0-alpha.4_linux_amd64 kubebuilder && mv kubebuilder /usr/local/
RUN export PATH=$PATH:/usr/local/kubebuilder/bin
