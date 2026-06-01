.DEFAULT_GOAL := help
GO ?= go
BIN := bin
SERVICES := metering billing catalog

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: build
build: ## Build all service binaries into ./bin
	@mkdir -p $(BIN)
	@for svc in $(SERVICES); do \
		echo "building $$svc"; \
		$(GO) build -o $(BIN)/$$svc ./cmd/$$svc || exit 1; \
	done

.PHONY: test
test: ## Run all tests
	$(GO) test ./... -count=1

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

.PHONY: check
check: vet test ## Vet + test (CI gate)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN)
