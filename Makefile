.PHONY: build test clean install run help lint deps-check setup

# Variables
BINARY_NAME := dl
BIN_DIR := bin
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
VERSION := $(shell cat VERSION 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildDate=$(BUILD_DATE)'"
GOINSTALL := go install

# Determine install directory using Go standard locations
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
  GOPATH ?= $(shell go env GOPATH)
  ifeq ($(GOPATH),)
    INSTALL_DIR := $(HOME)/go/bin
  else
    INSTALL_DIR := $(GOPATH)/bin
  endif
else
  INSTALL_DIR := $(GOBIN)
endif

build: ## Build the binary
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/dl
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/dl
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/dl
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/dl
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/dl
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/dl
	@echo "Build complete for all platforms"

build-release-tool: ## Build the release tool (developers only)
	@echo "Building release tool (for developers)..."
	go build -o release ./cmd/release
	@echo "Build complete: ./release"
	@echo "Note: This is a development tool and is not installed with 'go install'"

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "All checks passed!"

clean: ## Remove built binaries
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f dl-*
	rm -f release
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

# Development tools (added from sc_router patterns)
deps-check: ## Check dependencies and show version information
	@echo "Checking dependencies..."
	@echo "Go version:"
	go version
	@echo ""
	@echo "Required tools:"
	@echo "  - golangci-lint (preferred): go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@echo "  - golint (fallback): go install golang.org/x/lint/golint@latest"

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install: build ## Install binary using Go's standard installation method
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@$(GOINSTALL) $(LDFLAGS) ./cmd/dl
	@echo "Installation complete. Binary location: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "$$PATH" | grep -q "$(INSTALL_DIR)" || echo "Warning: $(INSTALL_DIR) is not in your PATH."

lint: ## Run golint (requires golint to be installed)
	@echo "Running golint..."
	golint ./...

run: build ## Build and run with example URL
	@echo "Running $(BINARY_NAME)..."
	bin/$(BINARY_NAME) -h

setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "Development environment ready"

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -v
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	go tool cover -func=coverage.out

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race ./...

version: ## Show the current version
	@echo "Version: $(VERSION)"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
