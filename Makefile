.PHONY: all build test clean proto sqlc

VERSION ?= dev
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/ferrarinobrakes/unofficial-valorant-api/internal/version.Version=$(VERSION) \
           -X github.com/ferrarinobrakes/unofficial-valorant-api/internal/version.GitCommit=$(GIT_COMMIT) \
           -X github.com/ferrarinobrakes/unofficial-valorant-api/internal/version.BuildTime=$(BUILD_TIME)

all: proto sqlc build

proto:
	buf generate

sqlc:
	sqlc generate

build:
	go build -ldflags "$(LDFLAGS)" -o bin/master.exe ./cmd/master
	go build -ldflags "$(LDFLAGS)" -o bin/client.exe ./cmd/client

test:
	go test -v -race -cover ./...

clean:
	rm -rf bin/
	rm -rf gen/

run-master:
	./bin/master.exe

run-client:
	./bin/client.exe
