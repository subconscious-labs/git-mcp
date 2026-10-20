BINARY      := git-mcp
PKG         := ./...
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X main.version=$(VERSION)

.PHONY: all build run test lint security-test ci clean

all: build

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/git-mcp

run: build
	./bin/$(BINARY)

test:
	go test -race -count=1 $(PKG)

security-test:
	go test -race -count=1 -run 'TestCVE|TestReject|TestRedact|TestChain' ./internal/security/...

lint:
	golangci-lint run ./...

ci: lint security-test test

clean:
	rm -rf bin dist coverage.txt
