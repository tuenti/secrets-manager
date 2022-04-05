# Build the manager binary
FROM golang:1.16 as builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${SECRETS_MANAGER_VERSION}" -a -o secrets-manager main.go


#
# Prod image
#

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as release
WORKDIR /
COPY --from=builder /workspace/secrets-manager .
USER 65532:65532

ENTRYPOINT ["/secrets-manager"]


#
# Dev image
#

FROM builder as dev

ENV ENVTEST_ASSETS_DIR=testbin
ENV ENVTEST_K8S_VERSION=1.19.2
# kubebuilder needed to run tests in development environment
RUN curl -L -O https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.1.0/kubebuilder_linux_amd64
RUN mv kubebuilder_linux_amd64 kubebuilder \
    && chmod 755 kubebuilder \
    && mv kubebuilder /usr/local/bin
RUN export PATH=$PATH:/usr/local/bin
