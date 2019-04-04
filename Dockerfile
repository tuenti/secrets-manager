# Stage 0
# Build binary file
FROM golang:1.11.5-alpine as builder
ENV GLIDE_VERSION v0.13.2

RUN wget "https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz" \
    && tar xf glide-${GLIDE_VERSION}-linux-amd64.tar.gz \
    && cp linux-amd64/glide $GOPATH/bin/ \
    && chmod +x $GOPATH/bin/glide

RUN apk add --update git make

ARG PROJECT_SLUG=github.com/tuenti/secrets-manager
COPY glide.yaml /go/src/$PROJECT_SLUG/glide.yaml
COPY glide.lock /go/src/$PROJECT_SLUG/glide.lock
WORKDIR /go/src/$PROJECT_SLUG
RUN glide install

COPY . /go/src/$PROJECT_SLUG
ARG SECRETS_MANAGER_VERSION
RUN make build-linux

# Stage 1
# Build actual docker image
FROM alpine:3.8
ARG PROJECT_SLUG=github.com/tuenti/secrets-manager
LABEL maintainer="sre@tuenti.com"
COPY --from=builder /go/src/$PROJECT_SLUG/build/secrets-manager /secrets-manager
RUN apk update
RUN apk add ca-certificates
ENTRYPOINT ["/secrets-manager"]
