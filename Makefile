# Jurigen Legal Case Context Builder
.PHONY: help build test clean swagger swagger-serve

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

# Test with verbose output
test-verbose: ## Run tests with verbose output
	go test -v ./...

# Clean build artifacts
clean: ## Clean build artifacts and generated files
	rm -rf bin/
	rm -rf docs/

# Install swagger generation tool
swagger-install: ## Install swag CLI tool
	go install github.com/swaggo/swag/cmd/swag@latest

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

# Lint code
lint: ## Run linter
	golangci-lint run

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
dev: clean test swagger ## Full development build: clean, test, and generate docs

# Check if required tools are installed
check-deps: ## Check if required development tools are installed
	@echo "🔍 Checking development dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
	@echo "✅ Go is installed: $$(go version)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "⚠️  golangci-lint not installed (optional for linting)"; }
	@command -v python3 >/dev/null 2>&1 || { echo "⚠️  python3 not installed (needed for swagger-serve)"; }
	@echo "✅ Development environment ready"
