.PHONY: all build install clean test test-cov lint fmt vet generate openapi-pull tidy run release-snapshot help

GO          ?= go
BINARY      ?= zd
DIST        ?= ./dist
PKG         := github.com/hackath0r/zd-cli
LDFLAGS     := -s -w \
	-X $(PKG)/internal/version.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) \
	-X $(PKG)/internal/version.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) \
	-X $(PKG)/internal/version.Date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

all: lint test build

build: ## Build the zd binary into the current directory
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/zd

install: ## Install zd to GOBIN
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" ./cmd/zd

generate: ## Regenerate the OpenAPI client from api/openapi.yaml
	$(GO) generate ./internal/zenduty/...

openapi-pull: ## Pull the latest OpenAPI spec from apidocs.zenduty.com
	@curl -fsSL https://apidocs.zenduty.com/openapi.json -o api/openapi.json
	@$(GO) run ./internal/tools/json2yaml api/openapi.json api/openapi.yaml
	@rm api/openapi.json
	@echo "OpenAPI spec refreshed; run 'make generate' next."

test: ## Run unit tests
	$(GO) test -race -count=1 ./...

test-cov: ## Run tests with coverage report
	$(GO) test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...
	$(GO) tool cover -html=coverage.txt -o coverage.html

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { echo >&2 "golangci-lint not installed; see https://golangci-lint.run/"; exit 1; }
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w -local $(PKG) . || true

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## go mod tidy
	$(GO) mod tidy

release-snapshot: ## Build a local goreleaser snapshot
	@command -v goreleaser >/dev/null 2>&1 || { echo >&2 "goreleaser not installed"; exit 1; }
	goreleaser release --clean --snapshot --skip=publish

clean: ## Remove build artifacts
	rm -rf $(DIST) $(BINARY) ximr coverage.txt coverage.html

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
