BINARY      := portkill
PKG         := ./cmd/portkill
VERSION_PKG := github.com/khanalsaroj/portkill/internal/version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).Date=$(DATE)

.PHONY: all build install test cover race vet fmt lint tidy clean help

all: build

build: ## Build the binary for the host platform into ./bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

install: ## Install portkill into $(go env GOPATH)/bin
	go install -trimpath -ldflags "$(LDFLAGS)" $(PKG)

test: ## Run the test suite
	go test ./...

cover: ## Run tests and print a coverage summary
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

race: ## Run tests with the race detector (needs CGO)
	CGO_ENABLED=1 go test -race ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go source
	gofmt -w .

lint: ## Run golangci-lint (install: https://golangci-lint.run)
	golangci-lint run

tidy: ## Tidy go.mod / go.sum
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin dist coverage.out

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
