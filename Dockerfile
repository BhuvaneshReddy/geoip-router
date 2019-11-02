FROM gcr.io/gcp-runtimes/go1-builder:1.13 as builder

WORKDIR /go/src/app
COPY cmd cmd
COPY pkg pkg

ENV CGO_ENABLED 0
ARG cmd=server

RUN BUILD_DATE=$(date -u +%d%m%Y.%H%M%S) && \
    DATE=$(date -u +%d%m%Y.%H%M%S) && \
    GO_VERSION=$(/usr/local/go/bin/go  version | awk '{print $3}') && \
    APP_NAME=${cmd} && \
    /usr/local/go/bin/go build -tags debug -o /dist/server -v -i -ldflags="-X github.com/etherlabsio/pkg/version.buildDate=$DATE -X github.com/etherlabsio/pkg/version.appName=$APP_NAME -X github.com/etherlabsio/pkg/version.goVersion=$GO_VERSION -s -w" ./cmd/${cmd}

# Application image.
FROM gcr.io/distroless/base:latest

WORKDIR /app
COPY --from=builder /dist /usr/local/bin/
COPY resources resources/

CMD ["/usr/local/bin/server"]