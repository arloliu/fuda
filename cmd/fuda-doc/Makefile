# Makefile for fuda-doc CLI
# Usage: make [target]

# Configuration
BINARY_NAME     := fuda-doc
TEST_TIMEOUT    ?= 3m
LINT_TIMEOUT    ?= 3m
COVERAGE_DIR    := ./.coverage
COVERAGE_OUT    := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML   := $(COVERAGE_DIR)/coverage.html

# Linter configuration
LINTER_GOMOD          := -modfile=linter.go.mod
GOLANGCI_LINT_VERSION := 2.5.0

# Version (can be overridden: make build VERSION=1.2.3)
VERSION         ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS         := -ldflags "-s -w -X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

.PHONY: help build test test-quick coverage clean-test-results lint linter-update linter-version fmt vet clean ci

## help: Show this help message
help:
	@echo "Available targets:" && \
	grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

##@ Build

## build: Build the fuda-doc binary
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Built $(BINARY_NAME)"

##@ Testing

## test: Run all tests with race detection
test: clean-test-results
	@echo "Running tests..."
	@CGO_ENABLED=1 go test ./... -timeout=$(TEST_TIMEOUT) -race
	@echo "All tests passed!"

## test-quick: Run tests without race detection (fast)
test-quick: clean-test-results
	@echo "Running tests without race detection..."
	@CGO_ENABLED=0 go test ./... -short -timeout=$(TEST_TIMEOUT)

## clean-test-results: Clean test artifacts
clean-test-results:
	@rm -f test.log *.pprof
	@go clean -testcache
	@rm -rf $(COVERAGE_DIR)

## coverage: Run tests with coverage
coverage: clean-test-results
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@CGO_ENABLED=1 go test ./... -timeout=$(TEST_TIMEOUT) -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated at $(COVERAGE_HTML)"

##@ Code Quality

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

linter-update:
	@echo "Install/update linter tool..."
	@go get -tool $(LINTER_GOMOD) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	@go mod verify $(LINTER_GOMOD)

linter-version:
	@go tool $(LINTER_GOMOD) golangci-lint --version

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

##@ Cleanup

## clean: Clean all build artifacts and caches
clean: clean-test-results
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@go clean -cache

##@ CI/CD

## ci: Run all CI checks (lint, vet, test, coverage)
ci: lint vet test coverage
