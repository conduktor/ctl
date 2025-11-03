export SHELL := /bin/sh
export SHELLOPTS:=$(if $(SHELLOPTS),$(SHELLOPTS):)pipefail:errexit

GO_LINT_VERSION ?= v2.5.0
.ONESHELL:


.PHONY: help
help: ## Prints help for targets with comments
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build:	## Build the provider
	go build -o conduktor .

.PHONY: fmt
fmt: ## Run go fmt
	go fmt ./...

tools:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GO_LINT_VERSION)

.PHONY: lint
lint: tools ## Run Golang linters
	@echo "==> Run Golang CLI linter..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GO_LINT_VERSION) version
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GO_LINT_VERSION) config verify
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GO_LINT_VERSION) run

.PHONY: test
test: ## Run tests only (no setup or cleanup)
	go test ./... -v $(TESTARGS) -timeout 10m
	./scripts/test_final_exec.sh

.PHONY: install
install: ## Install required tools and dependencies
	@./scripts/install_dev_dependencies.sh

.PHONY: setup-hooks
setup-hooks: install ## Setup pre-commit hooks
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Run 'make install' first."; exit 1; }
	@echo "[*] installing pre-commit hooks..."
	@pre-commit install
