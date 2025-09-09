.PHONY: build test clean install lint help run-tests test-coverage check-deps fmt vet

BINARY_NAME=screenshot-tweets
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test ./... -v

test-short: ## Run tests in short mode (skip integration tests)
	@echo "Running tests in short mode..."
	@go test ./... -short -v

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test ./... -bench=.

lint: check-deps ## Run linter
	@echo "Running golangci-lint..."
	@golangci-lint run

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

check-deps: ## Check if required tools are installed
	@echo "Checking dependencies..."
	@which golangci-lint > /dev/null || (echo "golangci-lint is required. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	@go install ./cmd/$(BINARY_NAME)

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

run-example: build ## Run example with test data
	@echo "Running example with test data..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --dry-run --verbose --file testdata/sample-input.md

dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

ci: fmt vet lint test ## Run all CI checks
	@echo "All CI checks passed!"

release: clean ci build ## Build release binary
	@echo "Release build completed!"

.DEFAULT_GOAL := help