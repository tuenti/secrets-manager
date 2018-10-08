# Stage 0
# Build binary file
FROM instrumentisto/glide:0.13.1-go1.10

RUN apk add --update make

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
FROM alpine
ARG PROJECT_SLUG=github.com/tuenti/secrets-manager
COPY --from=0 /go/src/$PROJECT_SLUG/build/secrets-manager /secrets-manager
ENTRYPOINT ["/secrets-manager"]
