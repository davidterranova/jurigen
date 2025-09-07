# Jurigen Legal Case Context Builder
.PHONY: help build test test-coverage test-ci test-verbose clean swagger swagger-serve mocks mocks-clean generate lint lint-install fmt fmt-check vet deps check-deps dev server interactive dag-show workflow-branch workflow-commit workflow-pr workflow-full workflow-help

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the jurigen binary
	go build -o bin/jurigen main.go

# Test all packages
test: ## Run all tests
	go test ./...

# Test with coverage
test-coverage: ## Run tests with coverage report
	go test -cover ./...

# Test with coverage and race detection for CI
test-ci: ## Run tests with coverage, race detection, and coverage file output for CI
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Test with verbose output
test-verbose: ## Run tests with verbose output
	go test -v ./...

# Clean build artifacts
clean: ## Clean build artifacts and generated files
	rm -rf bin/
	rm -rf docs/

# Generate mocks for testing
mocks: ## Generate all mocks for interfaces using go:generate directives
	@echo "🔧 Generating mocks using go:generate directives..."
	@mkdir -p internal/usecase/testdata/mocks
	@mkdir -p internal/adapter/http/testdata/mocks
	go generate ./internal/usecase/
	go generate ./internal/adapter/http/
	@echo "✅ All mocks generated using go:generate directives"

# Generate all code (more idiomatic Go approach)
generate: ## Generate all code using go:generate directives (idiomatic Go way)
	@echo "🔧 Running go generate for all packages..."
	go generate ./...
	@echo "✅ Code generation complete"

# Clean generated mocks
mocks-clean: ## Remove generated mock files
	@echo "🧹 Cleaning generated mocks..."
	rm -rf test/mocks/
	rm -rf internal/usecase/testdata/mocks/
	rm -rf internal/adapter/http/testdata/mocks/
	@echo "✅ Mocks cleaned"

# Install swagger generation tool
swagger-install: ## Install swag CLI tool
	go install github.com/swaggo/swag/cmd/swag@latest

# Note: mockgen is run via go:generate directives, no separate installation needed

# Install golangci-lint tool
lint-install: ## Install golangci-lint CLI tool
	@echo "📦 Installing golangci-lint (latest version)..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest
	@echo "✅ golangci-lint installed"

# Run code linting
lint: ## Run golangci-lint on the codebase
	@echo "🔍 Running golangci-lint..."
	@export PATH=$$PATH:$(shell go env GOPATH)/bin && golangci-lint run --no-config --enable=errcheck,govet,ineffassign,staticcheck,unused,goconst,gocritic,gocyclo,misspell,nakedret,nestif,prealloc,unconvert,unparam,whitespace ./internal/dag ./internal/usecase ./internal/port ./internal/adapter/http ./pkg/... ./cmd
	@echo "✅ Linting completed"

# Generate OpenAPI/Swagger documentation
swagger: swagger-install ## Generate OpenAPI documentation from code annotations
	@echo "🔄 Generating OpenAPI/Swagger documentation..."
	export PATH=$$PATH:$(shell go env GOPATH)/bin && swag init --dir . --output docs/swagger --parseDependency --parseInternal
	@echo "✅ OpenAPI documentation generated in docs/swagger/"
	@echo "📄 Spec file: docs/swagger/swagger.json"
	@echo "📄 YAML file: docs/swagger/swagger.yaml"
	@echo "📱 To serve docs: make swagger-serve"

# Serve Swagger UI locally
swagger-serve: ## Serve Swagger UI locally (requires swagger generation first)
	@if [ ! -f docs/swagger/swagger.json ]; then \
		echo "❌ No swagger docs found. Run 'make swagger' first."; \
		exit 1; \
	fi
	@echo "🌐 Starting Swagger UI server on http://localhost:8081/swagger/"
	@echo "🔧 Press Ctrl+C to stop"
	@cd docs/swagger && python3 -m http.server 8081

# Format code
fmt: ## Format Go code
	go fmt ./...

# Check code formatting
fmt-check: ## Check if Go code is formatted correctly
	@echo "🔍 Checking Go code formatting..."
	@if [ "$(shell gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Go code is not formatted. Please run 'make fmt'"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "✅ Go code is properly formatted"

# Verify Go code
vet: ## Run go vet to check for suspicious constructs
	@echo "🔍 Running go vet..."
	go vet ./...
	@echo "✅ Go vet completed"

# Run the server
server: ## Start the HTTP API server
	go run main.go server

# Run interactive CLI
interactive: ## Start interactive DAG traversal (requires --dag flag)
	@echo "Usage: make interactive DAG_FILE=path/to/dag.json [CONTEXT=true]"
	@if [ -z "$(DAG_FILE)" ]; then \
		echo "❌ DAG_FILE is required. Example: make interactive DAG_FILE=dag.json"; \
		exit 1; \
	fi
	@if [ "$(CONTEXT)" = "true" ]; then \
		go run main.go interactive --dag $(DAG_FILE) --context; \
	else \
		go run main.go interactive --dag $(DAG_FILE); \
	fi

# Show DAG structure
dag-show: ## Show DAG structure (requires --dag flag)
	@echo "Usage: make dag-show DAG_FILE=path/to/dag.json"
	@if [ -z "$(DAG_FILE)" ]; then \
		echo "❌ DAG_FILE is required. Example: make dag-show DAG_FILE=dag.json"; \
		exit 1; \
	fi
	go run main.go dag --dag $(DAG_FILE)

# Development workflow  
dev: clean swagger generate lint test ## Full development build: clean, generate docs, generate code, lint, test

# Download dependencies and verify
deps: ## Download and verify dependencies
	@echo "📦 Downloading Go dependencies..."
	go mod download
	go mod verify
	@echo "✅ Dependencies verified"

# Check if required tools are installed
check-deps: ## Check if required development tools are installed
	@echo "🔍 Checking development dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
	@echo "✅ Go is installed: $$(go version)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "✅ golangci-lint is installed: $$(golangci-lint version --short)"; \
	else \
		echo "⚠️  golangci-lint not installed, run 'make lint-install'"; \
	fi
	@if command -v gh >/dev/null 2>&1; then \
		echo "✅ GitHub CLI is installed: $$(gh --version | head -1)"; \
	else \
		echo "⚠️  GitHub CLI not installed (optional for PR automation)"; \
	fi
	@echo "✅ Development dependency check completed"

##############################################################################
# Workflow Automation
##############################################################################

# Create a new feature branch following naming conventions
workflow-branch: ## Create a new feature branch (Usage: make workflow-branch TYPE=feature NAME=my-feature)
	@if [ -z "$(TYPE)" ] || [ -z "$(NAME)" ]; then \
		echo "❌ Usage: make workflow-branch TYPE=feature NAME=my-feature"; \
		echo "📋 Types: feature, bugfix, hotfix, refactor, docs, test"; \
		exit 1; \
	fi
	@./scripts/workflow-automation.sh branch $(TYPE) "$(NAME)"

# Auto-commit staged changes with conventional commit message
workflow-commit: ## Auto-create conventional commit for staged changes
	@./scripts/workflow-automation.sh commit

# Create PR with auto-generated description
workflow-pr: ## Create Pull Request with auto-generated description
	@./scripts/workflow-automation.sh pr create

# Generate PR template only
workflow-pr-template: ## Generate PR description template without creating PR
	@./scripts/workflow-automation.sh pr template

# Full automated workflow setup
workflow-full: ## Full workflow: create branch and setup (Usage: make workflow-full NAME=my-feature [TYPE=feature])
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Usage: make workflow-full NAME=my-feature [TYPE=feature]"; \
		exit 1; \
	fi
	@./scripts/workflow-automation.sh workflow "$(NAME)" "$(or $(TYPE),feature)"

# Show workflow help
workflow-help: ## Show detailed workflow automation help
	@./scripts/workflow-automation.sh help
