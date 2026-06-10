# SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

.PHONY: help build test lint clean install run docker-build docker-run dev

# Variables
BINARY_NAME_COMPACTOR=openchami-logq-compactor
BINARY_NAME_QUERY=openchami-logq-query
GO=go
GOFLAGS=-v
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build both applications
	@echo "Building compactor..."
	cd compactor && $(GO) build $(GOFLAGS) $(LDFLAGS) -o ../bin/$(BINARY_NAME_COMPACTOR) .
	@echo "Building query..."
	cd query && $(GO) build $(GOFLAGS) $(LDFLAGS) -o ../bin/$(BINARY_NAME_QUERY) .

build-compactor: ## Build compactor only
	@echo "Building compactor..."
	cd compactor && $(GO) build $(GOFLAGS) $(LDFLAGS) -o ../bin/$(BINARY_NAME_COMPACTOR) .

build-query: ## Build query only
	@echo "Building query..."
	cd query && $(GO) build $(GOFLAGS) $(LDFLAGS) -o ../bin/$(BINARY_NAME_QUERY) .

test: ## Run tests for all modules
	@echo "Testing compactor..."
	cd compactor && $(GO) test $(GOFLAGS) -race -coverprofile=../coverage-compactor.out -covermode=atomic ./...
	@echo "Testing query..."
	cd query && $(GO) test $(GOFLAGS) -race -coverprofile=../coverage-query.out -covermode=atomic ./...

test-compactor: ## Run tests for compactor
	cd compactor && $(GO) test $(GOFLAGS) -race -coverprofile=../coverage-compactor.out -covermode=atomic ./...

test-query: ## Run tests for query
	cd query && $(GO) test $(GOFLAGS) -race -coverprofile=../coverage-query.out -covermode=atomic ./...

test-coverage: test ## Run tests with coverage report
	$(GO) tool cover -html=coverage-compactor.out -o coverage-compactor.html
	$(GO) tool cover -html=coverage-query.out -o coverage-query.html
	@echo "Coverage reports generated: coverage-compactor.html, coverage-query.html"

lint: ## Run golangci-lint on all modules
	@echo "Linting compactor..."
	cd compactor && golangci-lint run
	@echo "Linting query..."
	cd query && golangci-lint run

lint-compactor: ## Run golangci-lint on compactor
	cd compactor && golangci-lint run

lint-query: ## Run golangci-lint on query
	cd query && golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix on all modules
	@echo "Linting compactor with auto-fix..."
	cd compactor && golangci-lint run --fix
	@echo "Linting query with auto-fix..."
	cd query && golangci-lint run --fix

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage-*.out coverage-*.html
	cd compactor && $(GO) clean -cache
	cd query && $(GO) clean -cache

install: ## Install dependencies for all modules
	@echo "Installing dependencies for compactor..."
	cd compactor && $(GO) mod download && $(GO) mod verify
	@echo "Installing dependencies for query..."
	cd query && $(GO) mod download && $(GO) mod verify

tidy: ## Tidy go.mod for all modules
	@echo "Tidying compactor..."
	cd compactor && $(GO) mod tidy
	@echo "Tidying query..."
	cd query && $(GO) mod tidy

dev: clean build ## Clean and build binaries

run-compactor: build-compactor ## Build and run compactor
	./bin/$(BINARY_NAME_COMPACTOR)

run-query: build-query ## Build and run query
	./bin/$(BINARY_NAME_QUERY)

docker-build: ## Build Docker images
	docker build -t $(BINARY_NAME_COMPACTOR):latest -f compactor/Dockerfile .
	docker build -t $(BINARY_NAME_QUERY):latest -f query/Dockerfile .

docker-build-compactor: ## Build compactor Docker image
	docker build -t $(BINARY_NAME_COMPACTOR):latest -f compactor/Dockerfile .

docker-build-query: ## Build query Docker image
	docker build -t $(BINARY_NAME_QUERY):latest -f query/Dockerfile .

docker-run-compactor: docker-build-compactor ## Build and run compactor container
	docker run --rm $(BINARY_NAME_COMPACTOR):latest

docker-run-query: docker-build-query ## Build and run query container
	docker run --rm $(BINARY_NAME_QUERY):latest

release-snapshot: ## Create a snapshot release with GoReleaser
	goreleaser release --snapshot --clean

fmt: ## Format code
	cd compactor && $(GO) fmt ./...
	cd query && $(GO) fmt ./...
	goimports -w compactor query

vet: ## Run go vet
	cd compactor && $(GO) vet ./...
	cd query && $(GO) vet ./...

vuln: ## Check for vulnerabilities
	cd compactor && govulncheck ./...
	cd query && govulncheck ./...

reuse: ## Check REUSE compliance
	reuse lint

reuse-spdx: ## Generate SPDX bill of materials
	reuse spdx -o reuse.spdx

reuse-install: ## Install REUSE tool
	@command -v pipx >/dev/null 2>&1 || { echo "pipx is required but not installed. Install it with: python3 -m pip install --user pipx"; exit 1; }
	pipx install reuse
	@echo "REUSE tool installed successfully"

reuse-annotate: ## Add REUSE headers to all files in the repository
	@echo "Annotating files with REUSE headers..."
	@echo "This will add SPDX headers to files that don't have them yet."
# REUSE-IgnoreStart
	@read -p "Copyright holder [OpenCHAMI a Series of LF Projects, LLC]: " holder; \
	holder=$${holder:-OpenCHAMI a Series of LF Projects, LLC}; \
	read -p "License [MIT]: " license; \
	license=$${license:-MIT}; \
	read -p "Year [$(shell date +%Y)]: " year; \
	year=$${year:-$(shell date +%Y)}; \
	echo "Annotating with: SPDX-FileCopyrightText: Copyright © $$year $$holder"; \
	echo "                 SPDX-License-Identifier: $$license"; \
	reuse annotate --copyright="$$holder" --license="$$license" --year="$$year" --skip-existing --recursive --skip-unrecognized .
# REUSE-IgnoreEnd

reuse-download-license: ## Download a license file (usage: make reuse-download-license LICENSE=MIT)
	@if [ -z "$(LICENSE)" ]; then \
		echo "Error: LICENSE variable is required. Usage: make reuse-download-license LICENSE=MIT"; \
		exit 1; \
	fi
	reuse download $(LICENSE)

pre-commit-install: ## Install pre-commit tool
	@command -v pipx >/dev/null 2>&1 || { echo "pipx is required but not installed. Install it with: python3 -m pip install --user pipx"; exit 1; }
	pipx install pre-commit
	@echo "pre-commit installed successfully"

pre-commit-setup: ## Install pre-commit hooks
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit is not installed. Run 'make pre-commit-install' first."; exit 1; }
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "pre-commit hooks installed successfully"

pre-commit-run: ## Run pre-commit hooks on all files
	pre-commit run --all-files

pre-commit-update: ## Update pre-commit hooks to latest versions
	pre-commit autoupdate

setup-dev: reuse-install pre-commit-install pre-commit-setup ## Set up development environment (install tools and hooks)
	@echo ""
	@echo "Development environment setup complete!"
	@echo "Next steps:"
	@echo "  1. Run 'make reuse-annotate' to add REUSE headers to all files"
	@echo "  2. Run 'make pre-commit-run' to test pre-commit hooks"
	@echo "  3. Start coding! Pre-commit hooks will run automatically on git commit"
	@echo ""
	@echo "Optional: Install 'act' to test GitHub Actions locally:"
	@echo "  brew install act"
	@echo "  make act-list  # List available workflows"

act-install: ## Install act (GitHub Actions local runner) via Homebrew
	@command -v brew >/dev/null 2>&1 || { echo "Homebrew is required. Install from https://brew.sh"; exit 1; }
	brew install act
	@echo "act installed successfully"

act-list: ## List all GitHub Actions workflows
	@command -v act >/dev/null 2>&1 || { echo "act is not installed. Run 'make act-install' first."; exit 1; }
	@echo "Available workflows:"
	@ls -1 .github/workflows/*.yaml .github/workflows/*.yml 2>/dev/null | sed 's/.*\//  - /' || echo "  No workflows found"

act-lint: ## Run GitHub Actions golangci-lint workflow locally
	@command -v act >/dev/null 2>&1 || { echo "act is not installed. Run 'make act-install' first."; exit 1; }
	act push -W .github/workflows/golangci-lint.yaml --container-architecture linux/amd64

act-reuse: ## Run GitHub Actions REUSE workflow locally
	@command -v act >/dev/null 2>&1 || { echo "act is not installed. Run 'make act-install' first."; exit 1; }
	act push -W .github/workflows/REUSE.yaml --container-architecture linux/amd64

all: clean install lint test build ## Run all checks and build

.DEFAULT_GOAL := help
