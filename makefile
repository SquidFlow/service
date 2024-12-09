GO=go
GOFLAGS=-v

VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD)
INSTALLATION_MANIFESTS_URL := "github.com/squidflow/service/manifests/base"
INSTALLATION_MANIFESTS_THIRD_PARTY := "github.com/squidflow/service/manifests/third-party"

LDFLAGS := -X 'github.com/squidflow/service/pkg/store.version=$(VERSION)' \
           -X 'github.com/squidflow/service/pkg/store.buildDate=$(BUILD_DATE)' \
           -X 'github.com/squidflow/service/pkg/store.gitCommit=$(GIT_COMMIT)' \
           -X 'github.com/squidflow/service/pkg/store.installationManifestsURL=github.com/squidflow/service/manifests/base' \
           -X 'github.com/squidflow/service/pkg/store.installationManifestsThirdParty=github.com/squidflow/service/manifests/third-party'

# Default target
.DEFAULT_GOAL := build

# Build application
.PHONY: build
build: build-service build-supervisor

build-service:
	go build -v -o output/service -ldflags "$(LDFLAGS)" cmd/service/service.go

build-supervisor:
	go build -v -o output/supervisor -ldflags "$(LDFLAGS)" cmd/supervisor/supervisor.go

.PHONY: test
test:
	$(GO) test ./...

.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

.PHONY: all
all: clean build test

.PHONY: codegen
codegen: $(GOBIN)/mockgen $(GOBIN)/interfacer
	interfacer -for github.com/go-git/go-git/v5.Repository -as gogit.Repository -o pkg/git/gogit/repo.go
	interfacer -for github.com/go-git/go-git/v5.Worktree -as gogit.Worktree -o pkg/git/gogit/worktree.go
	interfacer -for github.com/google/go-github/v43/github.UsersService -as github.Users -o pkg/git/github/users.go
	interfacer -for github.com/google/go-github/v43/github.RepositoriesService -as github.Repositories -o pkg/git/github/repos.go
	interfacer -for github.com/google/go-github/v43/github.PullRequestsService -as github.PullRequests -o pkg/git/github/pull_requests.go
	go generate ./...

$(GOBIN)/mockgen:
	@go install github.com/golang/mock/mockgen@v1.6.0
	@mockgen -version

$(GOBIN)/interfacer:
	@echo "Installing interfacer..."
	@go install github.com/rjeczalik/interfaces/cmd/interfacer@latest

$(GOBIN)/golangci-lint:
	@mkdir dist || true
	@echo installing: golangci-lint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.55.2

# Docker ENV
DOCKER_REGISTRY ?= wangguohao
IMAGE_NAME ?= h4-service
IMAGE_TAG ?= $(VERSION)

.PHONY: docker-build
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: docker-all
docker-all: docker-build docker-push

# Help information
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make build           - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make all            - Clean, build, and test"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-push    - Push Docker image to registry"
	@echo "  make docker-all     - Build and push Docker image"
