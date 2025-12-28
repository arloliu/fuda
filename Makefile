# Makefile for mebo project
# Usage: make [target]

# Configuration
TEST_TIMEOUT    ?= 3m
LINT_TIMEOUT    ?= 3m
COVERAGE_DIR    := ./.coverage
COVERAGE_OUT    := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML   := $(COVERAGE_DIR)/coverage.html

# Source files
ALL_GO_FILES    := $(shell find . -name "*.go")
LATEST_GIT_TAG       := $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || echo "v0.0.0")
LATEST_VAULT_GIT_TAG := $(shell git describe --tags --abbrev=0 --match 'vault/v*' 2>/dev/null | sed 's|^vault/||' || echo "v0.0.0")

# Linter configuration
LINTER_GOMOD          := -modfile=linter.go.mod
GOLANGCI_LINT_VERSION := 2.5.0

# Default target
.DEFAULT_GOAL := help

.PHONY: help test test-vault test-quick coverage clean-test-results lint fmt vet clean gomod-tidy update-pkg-cache ci

## help: Show this help message
help:
	@echo "Available targets:" && \
	grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

## test: Run all tests (unit + integration + vault)
test: clean-test-results
	@echo "Running tests..."
	@echo "  -> fuda (root module)"
	@CGO_ENABLED=1 go test ./... -timeout=$(TEST_TIMEOUT) -race
	@echo "  -> fuda/vault"
	@cd vault && CGO_ENABLED=1 go test ./... -timeout=$(TEST_TIMEOUT) -race
	@echo "All tests passed!"

## test-vault: Run only vault package tests
test-vault: clean-test-results
	@echo "Running vault tests..."
	@cd vault && CGO_ENABLED=1 go test ./... -v -timeout=$(TEST_TIMEOUT) -race

## test-quick: Run tests without race detection (fast)
test-quick: clean-test-results
	@echo "Running tests without race detection..."
	@CGO_ENABLED=0 go test ./... -short -timeout=$(TEST_TIMEOUT)
	@cd vault && CGO_ENABLED=0 go test ./... -short -timeout=$(TEST_TIMEOUT)

## clean-test-results: Clean test artifacts
## clean-test-results: Clean test artifacts
clean-test-results:
	@rm -f test.log *.pprof
	@go clean -testcache
	@rm -rf $(COVERAGE_DIR)

## coverage: Run tests with coverage
coverage: clean-test-results
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@echo "  -> fuda (root module)"
	@CGO_ENABLED=1 go test ./... -timeout=$(TEST_TIMEOUT) -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated at $(COVERAGE_HTML)"

##@ Code Quality

.PHONY: linter-update linter-version
linter-update:
	@echo "Install/update linter tool..."
	@go get -tool $(LINTER_GOMOD) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	@go mod verify $(LINTER_GOMOD)

linter-version:
	@go tool $(LINTER_GOMOD) golangci-lint --version

## lint: Run linters
lint:
	@echo "Checking golangci-lint version..."
	@INSTALLED_VERSION=$$(go tool $(LINTER_GOMOD) golangci-lint --version 2>/dev/null | grep -oE 'version [^ ]+' | cut -d' ' -f2 || echo "not-installed"); \
	if [ "$$INSTALLED_VERSION" = "not-installed" ]; then \
		echo "Error: golangci-lint not found. Run 'make linter-update' to install."; \
		exit 1; \
	elif [ "$$INSTALLED_VERSION" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "Warning: golangci-lint version mismatch!"; \
		echo "  Expected: $(GOLANGCI_LINT_VERSION)"; \
		echo "  Installed: $$INSTALLED_VERSION"; \
		echo "  Run 'make linter-update' to install the correct version."; \
		exit 1; \
	else \
		echo "✓ golangci-lint $(GOLANGCI_LINT_VERSION) is installed"; \
	fi
	@echo "Running linters..."
	@go tool $(LINTER_GOMOD) golangci-lint run --timeout=$(LINT_TIMEOUT)

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w $(ALL_GO_FILES)

## vet: Run go vet
## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@cd vault && go vet ./...

##@ Build & Dependencies

## gomod-tidy: Tidy go.mod and go.sum for all modules
gomod-tidy:
	@echo "Tidying go modules..."
	@echo "  -> fuda (root)"
	@go mod tidy
	@go mod verify
	@echo "  -> fuda/vault"
	@cd vault && go mod tidy && go mod verify

## update-pkg-cache: Update Go package cache with latest git tags
update-pkg-cache:
	@echo "Updating package cache..."
	@echo "  -> fuda $(LATEST_GIT_TAG)"
	@curl -sf https://proxy.golang.org/github.com/arloliu/fuda/@v/$(LATEST_GIT_TAG).info > /dev/null || \
		echo "Warning: Failed to update fuda $(LATEST_GIT_TAG) package cache"
	@echo "  -> fuda/vault $(LATEST_VAULT_GIT_TAG)"
	@curl -sf https://proxy.golang.org/github.com/arloliu/fuda/vault/@v/$(LATEST_VAULT_GIT_TAG).info > /dev/null || \
		echo "Warning: Failed to update vault $(LATEST_VAULT_GIT_TAG) package cache"

##@ Cleanup

## clean: Clean all build artifacts and caches
clean: clean-test-results
	@echo "Cleaning build artifacts..."
	@go clean -cache -modcache -i -r
	@rm -rf dist/ bin/

##@ CI/CD

## ci: Run all CI checks (lint, test, coverage)
ci: lint vet test coverage
