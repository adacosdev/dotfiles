.PHONY: build build-dev test test-cover lint fmt tidy clean install publish

BINARY=adacosdev-dots
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

build: ## Build the binary
	go build -ldflags="-s -w" -o bin/$(BINARY) ./cmd/adacosdev-dots

build-dev: ## Build with debug info
	go build -o bin/$(BINARY) ./cmd/adacosdev-dots

test: ## Run all tests
	go test ./... -v

test-cover: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...

tidy: ## Tidy go modules
	go mod tidy

clean: ## Clean build artifacts
	rm -rf bin/

install: build ## Install to ~/.local/bin
	install -Dm755 bin/$(BINARY) ~/.local/bin/$(BINARY)

publish: build ## Build for multiple platforms
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY)-linux-amd64 ./cmd/adacosdev-dots
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/$(BINARY)-linux-arm64 ./cmd/adacosdev-dots
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY)-darwin-amd64 ./cmd/adacosdev-dots
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/$(BINARY)-darwin-arm64 ./cmd/adacosdev-dots

.DEFAULT_GOAL := build
