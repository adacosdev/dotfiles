# Integration tests for adacosdev-dots TUI
# Requires the binary to be built first

.PHONY: all test build help clean

# Binary name and path
BINARY_NAME := adacosdev-dots
BINARY_PATH := ../../bin/$(BINARY_NAME)

all: test

# Build the binary if not exists
build:
	@echo "Building $(BINARY_NAME)..."
	@cd ../.. && go build -o bin/$(BINARY_NAME) ./cmd/adacosdev-dots

# Run integration tests
test: build
	@echo "Running integration tests..."
	@cd ../.. && go test -v ./tests/integration/...

# Run integration tests with coverage
test-cover: build
	@echo "Running integration tests with coverage..."
	@cd ../.. && go test -v -cover ./tests/integration/...

# Run specific integration test
test-run: build
	@echo "Running $(TEST_NAME)..."
	@cd ../.. && go test -v -run "$(TEST_NAME)" ./tests/integration/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f ../../bin/$(BINARY_NAME)
	@cd ../.. && go clean -testcache

# Show test output without rebuilding
test-quick:
	@echo "Running integration tests (quick - no rebuild)..."
	@cd ../.. && go test -v ./tests/integration/...

# Build and show detailed output
test-verbose: build
	@echo "Running integration tests (verbose)..."
	@cd ../.. && go test -v -x ./tests/integration/...

# Run with race detector
test-race: build
	@echo "Running integration tests with race detector..."
	@cd ../.. && go test -v -race ./tests/integration/...

help:
	@echo "Integration tests for adacosdev-dots"
	@echo ""
	@echo "Targets:"
	@echo "  all          Run tests (default)"
	@echo "  build        Build the binary"
	@echo "  test         Build and run tests"
	@echo "  test-cover   Build and run tests with coverage"
	@echo "  test-run     Run specific test (set TEST_NAME=TestName)"
	@echo "  test-quick   Run tests without rebuilding"
	@echo "  test-verbose Build and run with verbose output"
	@echo "  test-race    Run tests with race detector"
	@echo "  clean        Remove build artifacts"
	@echo "  help         Show this help"
