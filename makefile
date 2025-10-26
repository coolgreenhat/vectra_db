# VectraDB Makefile

# Variables
BINARY_NAME=vectordbd
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/vectordbd
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
VERSION=$(shell git describe --tags --always --dirty)

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_PATH)

# Run with hot reload (requires air)
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@air

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./... -v -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report generated: coverage.html"

# Run tests with cleanup
.PHONY: test-clean
test-clean: clean-test-dbs test

# Clean test database files
.PHONY: clean-test-dbs
clean-test-dbs:
	@echo "Cleaning up test database files..."
	@find . -name "test_*.db" -type f -delete 2>/dev/null || true
	@echo "Test database files cleaned up"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t vectradb:$(VERSION) .
	@docker build -t vectradb:latest .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 vectradb:latest

# Docker compose up
.PHONY: docker-up
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

# Docker compose down
.PHONY: docker-down
docker-down:
	@echo "Stopping services with Docker Compose..."
	@docker-compose down

# Generate API documentation
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@swag init -g cmd/vectordbd/main.go -o docs/swagger

# Security scan
.PHONY: security
security:
	@echo "Running security scan..."
	@gosec ./...

# Check for vulnerabilities
.PHONY: vuln
vuln:
	@echo "Checking for vulnerabilities..."
	@govulncheck ./...

# Generate mocks
.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	@mockgen -source=internal/store/interfaces.go -destination=tests/mocks/store_mock.go

# Database operations
.PHONY: db-backup
db-backup:
	@echo "Backing up database..."
	@cp vectra.db vectra.db.backup.$(shell date +%Y%m%d_%H%M%S)

# Database restore
.PHONY: db-restore
db-restore:
	@echo "Restoring database..."
	@cp vectra.db.backup.$(shell ls vectra.db.backup.* | tail -1 | cut -d. -f3) vectra.db

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode with hot reload"
	@echo "  test          - Run tests"
	@echo "  test-clean    - Run tests with cleanup"
	@echo "  clean-test-dbs - Clean test database files"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-up     - Start services with Docker Compose"
	@echo "  docker-down   - Stop services with Docker Compose"
	@echo "  docs          - Generate API documentation"
	@echo "  security      - Run security scan"
	@echo "  vuln          - Check for vulnerabilities"
	@echo "  mocks         - Generate mocks"
	@echo "  db-backup     - Backup database"
	@echo "  db-restore    - Restore database"
	@echo "  help          - Show this help"
