# Epic Games Free Games Discord Bot - Makefile

.PHONY: build run test clean help install-deps

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-deps - Install Go dependencies"
	@echo "  lint         - Run linter (requires golangci-lint)"

# Build the application
build:
	@echo "Building Epic Games Discord Bot..."
	go build -o bin/epic-games-bot cmd/bot/main.go

# Run the application
run:
	@echo "Running Epic Games Discord Bot..."
	go run cmd/bot/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run linter (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	golangci-lint run

# Create binary directory
bin:
	mkdir -p bin