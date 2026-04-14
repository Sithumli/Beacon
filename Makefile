.PHONY: all build clean test run run-echo run-code help

# Variables
BINARY_SERVER=a2a-server
BINARY_CLI=a2a

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

all: build

## build: Build all binaries
build: build-server build-cli build-agents

build-server:
	@echo "Building server..."
	$(GOBUILD) -o bin/$(BINARY_SERVER) ./cmd/server

build-cli:
	@echo "Building CLI..."
	$(GOBUILD) -o bin/$(BINARY_CLI) ./cmd/a2a

build-agents:
	@echo "Building example agents..."
	$(GOBUILD) -o bin/echo-agent ./examples/echo-agent
	$(GOBUILD) -o bin/code-agent ./examples/code-agent

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f a2a.db

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## run: Run the server
run: build-server
	@echo "Running server..."
	./bin/$(BINARY_SERVER)

## run-memory: Run server with in-memory storage
run-memory: build-server
	@echo "Running server with in-memory storage..."
	./bin/$(BINARY_SERVER) --memory

## run-echo: Build and run echo agent
run-echo: build-agents
	@echo "Running echo agent..."
	./bin/echo-agent

## run-code: Build and run code agent
run-code: build-agents
	@echo "Running code agent..."
	./bin/code-agent

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t a2a-platform:latest .

## docker-run: Run with Docker Compose
docker-run:
	docker-compose up -d

## docker-stop: Stop Docker Compose
docker-stop:
	docker-compose down

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .

## help: Show this help
help:
	@echo "A2A Platform - Available targets:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Default target
.DEFAULT_GOAL := help
