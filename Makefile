# Makefile for gohot

# Variables
BINARY_NAME=gohot
VERSION=$(shell grep 'var __VERSION__' gohot.go | sed 's/.*"\(.*\)".*/\1/')
BUILD_DIR=build
MAIN_FILE=gohot.go
GO_FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Build flags
LDFLAGS=-ldflags "-w -s -X main.__VERSION__=$(VERSION)"
LDFLAGS_DEV=-ldflags "-X main.__VERSION__=$(VERSION)"

# Platform targets
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Colors for output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.PHONY: all build build-dev clean test test-coverage install run fmt vet lint help \
	build-all version deps tidy check pre-commit bench test-verbose

# Default target
all: clean fmt vet test build

## help: Display this help message
help:
	@echo "$(CYAN)Available targets:$(NC)"
	@echo ""
	@echo "$(GREEN)Building:$(NC)"
	@echo "  make build          - Build optimized binary for current platform"
	@echo "  make build-dev      - Build binary with debug info"
	@echo "  make build-all      - Build binaries for all platforms"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo ""
	@echo "$(GREEN)Testing:$(NC)"
	@echo "  make test           - Run tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make bench          - Run benchmarks"
	@echo ""
	@echo "$(GREEN)Code Quality:$(NC)"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run golangci-lint (requires installation)"
	@echo "  make check          - Run fmt, vet, and test"
	@echo "  make pre-commit     - Run all checks before committing"
	@echo ""
	@echo "$(GREEN)Maintenance:$(NC)"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download dependencies"
	@echo "  make tidy           - Tidy and verify dependencies"
	@echo "  make version        - Show version"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  make run            - Run the application"
	@echo "  make watch          - Run with hot reload (dogfooding)"
	@echo ""
	@echo "Current version: $(YELLOW)$(VERSION)$(NC)"

## build: Build optimized binary for current platform
build:
	@echo "$(CYAN)Building $(BINARY_NAME) v$(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-dev: Build binary with debug info (faster builds)
build-dev:
	@echo "$(CYAN)Building $(BINARY_NAME) (dev mode)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS_DEV) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✓ Dev build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Build binaries for all platforms
build-all: clean
	@echo "$(CYAN)Building for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} ; \
		output=$(BUILD_DIR)/$(BINARY_NAME)-v$(VERSION)-$$GOOS-$$GOARCH ; \
		if [ "$$GOOS" = "windows" ]; then output=$${output}.exe; fi ; \
		echo "Building for $$GOOS/$$GOARCH..." ; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o $$output $(MAIN_FILE) ; \
	done
	@echo "$(GREEN)✓ Cross-platform builds complete$(NC)"
	@ls -lh $(BUILD_DIR)

## install: Install binary to GOPATH/bin
install:
	@echo "$(CYAN)Installing $(BINARY_NAME)...$(NC)"
	go install $(LDFLAGS) .
	@echo "$(GREEN)✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

## clean: Remove build artifacts
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## test: Run tests
test:
	@echo "$(CYAN)Running tests...$(NC)"
	go test -v ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(CYAN)Running tests (verbose)...$(NC)"
	go test -v -race ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## bench: Run benchmarks
bench:
	@echo "$(CYAN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

## fmt: Format code with gofmt
fmt:
	@echo "$(CYAN)Formatting code...$(NC)"
	@gofmt -s -w $(GO_FILES)
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(CYAN)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet passed$(NC)"

## lint: Run golangci-lint (requires installation)
lint:
	@echo "$(CYAN)Running golangci-lint...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... ; \
		echo "$(GREEN)✓ Lint passed$(NC)" ; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed. Install: https://golangci-lint.run/usage/install/$(NC)" ; \
	fi

## check: Run fmt, vet, and test
check: fmt vet test

## pre-commit: Run all checks before committing
pre-commit: fmt vet lint test
	@echo "$(GREEN)✓ All pre-commit checks passed!$(NC)"

## deps: Download dependencies
deps:
	@echo "$(CYAN)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## tidy: Tidy and verify dependencies
tidy:
	@echo "$(CYAN)Tidying dependencies...$(NC)"
	go mod tidy
	go mod verify
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## version: Show version
version:
	@echo "$(BINARY_NAME) version: $(VERSION)"

## run: Run the application
run: build-dev
	@echo "$(CYAN)Running $(BINARY_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

## watch: Run with hot reload (dogfooding - using gohot to develop gohot!)
watch:
	@echo "$(CYAN)Running with hot reload (self-hosting!)...$(NC)"
	@if [ -f "./$(BINARY_NAME)" ]; then \
		./$(BINARY_NAME) ; \
	elif [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) ; \
	else \
		echo "$(YELLOW)Binary not found. Building first...$(NC)" ; \
		$(MAKE) build-dev ; \
		$(BUILD_DIR)/$(BINARY_NAME) ; \
	fi
