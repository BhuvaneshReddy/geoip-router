FROM golang:1.13 as builder

WORKDIR /go/src/app

# add Gopkg.toml and Gopkg.lock
COPY go.mod go.mod
COPY go.sum go.sum

# Force the go compiler to use modules
ENV GO111MODULE=on

# This is the ‘magic’ step that will download all the dependencies that are specified in
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the  go mod download
# command will _ only_ be re-run when the go.mod or go.sum file change
# (or when we add another docker instruction this line)
RUN go mod download

COPY cmd cmd
COPY pkg pkg

ENV CGO_ENABLED 0

RUN DATE=$(date -u +%d%m%Y.%H%M%S) && \
    GO_VERSION=$(go  version | awk '{print $3}') && \
    go build -tags debug -o /dist/server -v -i -ldflags="-X github.com/etherlabsio/pkg/version.buildDate=$DATE -X github.com/etherlabsio/pkg/version.appName=$APP_NAME -X github.com/etherlabsio/pkg/version.goVersion=$GO_VERSION -s -w" ./cmd/server

# Application image.
FROM gcr.io/distroless/base:latest

WORKDIR /app
COPY --from=builder /dist /usr/local/bin/
COPY resources resources/

CMD ["/usr/local/bin/server"]