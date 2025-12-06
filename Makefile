.PHONY: build test clean install run help go-install

BINARY_NAME=dl
BIN_DIR=bin
VERSION=$(shell cat VERSION)

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

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/dl
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

build-release-tool: ## Build the release tool (developers only)
	@echo "Building release tool (for developers)..."
	go build -o release ./cmd/release
	@echo "Build complete: ./release"
	@echo "Note: This is a development tool and is not installed with 'go install'"

install: build ## Install binary using Go's standard installation method
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/dl
	@echo "Installation complete. Binary location: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "$$PATH" | grep -q "$(INSTALL_DIR)" || echo "Warning: $(INSTALL_DIR) is not in your PATH."

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -cover ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race ./...

clean: ## Remove built binaries
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f dl-*
	rm -f release
	@echo "Clean complete"

run: build ## Build and run with example URL
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) -h

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Run golint (requires golint to be installed)
	@echo "Running golint..."
	golint ./...

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-linux-amd64 ./cmd/dl
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-linux-arm64 ./cmd/dl
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-darwin-amd64 ./cmd/dl
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-darwin-arm64 ./cmd/dl
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-windows-amd64.exe ./cmd/dl
	@echo "Build complete for all platforms"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

check: fmt vet test ## Run all checks (fmt, vet, test)
	@echo "All checks passed"

version: ## Show the current version
	@echo "Version: $(VERSION)"
