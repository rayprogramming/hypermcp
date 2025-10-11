# Makefile for hypermcp

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run all tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race: ## Run tests with race detector
	go test ./... -race -v

.PHONY: bench
bench: ## Run benchmarks
	go test ./... -bench=. -benchmem

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run --timeout=3m

.PHONY: lint-fix
lint-fix: ## Run golangci-lint and auto-fix issues
	golangci-lint run --fix --timeout=3m

.PHONY: fmt
fmt: ## Format code with gofmt
	gofmt -s -w .

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: build
build: ## Build the library (compile check)
	go build ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	go mod tidy
	go mod verify

.PHONY: mod-download
mod-download: ## Download go modules
	go mod download

.PHONY: clean
clean: ## Clean build artifacts and test cache
	go clean -testcache
	rm -f coverage.out coverage.html

.PHONY: check
check: fmt-check vet lint test ## Run all checks (fmt, vet, lint, test)

.PHONY: ci
ci: mod-download check test-coverage ## Run CI checks locally

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

.PHONY: pre-commit
pre-commit: fmt vet test ## Run pre-commit checks

.DEFAULT_GOAL := help
