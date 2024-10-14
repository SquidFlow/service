# Variable definitions
BINARY_NAME=server
GO=go
GOFLAGS=-v
MAIN_FILE=cmd/server/*.go

VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%H:%M:%S@%Y-%m-%d')
GIT_COMMIT := $(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Default target
.DEFAULT_GOAL := build

# Build application
.PHONY: build
build: build-service build-bootstrap

build-service:
	$(GO) build $(GOFLAGS) -o output/service $(LDFLAGS) cmd/service/service.go

build-bootstrap:
	$(GO) build $(GOFLAGS) -o output/bootstrap $(LDFLAGS) cmd/bootstrap/bootstrap.go

# Run application
.PHONY: run
run: build
	./$(BINARY_NAME) run

# Run tests
.PHONY: test
test:
	$(GO) test ./...

# Clean build artifacts
.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

# Execute all operations
.PHONY: all
all: clean build test

# Generate Swagger documentation (if using swag)
.PHONY: swagger
swagger:
	swag init -g $(MAIN_FILE) -o ./docs/swagger

# Format code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Check code style
.PHONY: lint
lint:
	golangci-lint run

# Helm deployment
.PHONY: deploy
deploy:
	helm upgrade --install application-api ./deploy/application-api

.PHONY: codegen
codegen: $(GOBIN)/mockgen
	rm -f docs/commands/*
	go generate ./...

$(GOBIN)/mockgen:
	@go install github.com/golang/mock/mockgen@v1.6.0
	@mockgen -version

$(GOBIN)/golangci-lint:
	@mkdir dist || true
	@echo installing: golangci-lint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.55.2

# Help information
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run the application"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make all      - Clean, build, and test"
