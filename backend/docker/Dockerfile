FROM golang:1.26-trixie

RUN set -e \
 && apt update \
 && apt install -y python3-watchdog rsync ffmpeg \
 && go install golang.org/x/tools/gopls@v0.21.1 \
 && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 \
 && go install github.com/nametake/golangci-lint-langserver@v0.0.12

WORKDIR /workdir/backend

COPY go.mod go.sum .
COPY cmd cmd
COPY internal internal

RUN set -e \
 && go test --tags "fts5" ./... # run tests to get an initial compilation
