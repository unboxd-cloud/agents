.DEFAULT_GOAL := help
GO ?= go
BIN := bin
# All commands under ./cmd are built; these get OCI images.
SERVICES := metering billing catalog compliance admin operator orgconsole
# Every buildable command (services + CLI + agents). agentql-wasm is built for
# GOOS=js only (see the agentql-wasm target), so it is excluded from the native
# build loop.
CMDS := $(filter-out agentql-wasm,$(notdir $(wildcard cmd/*)))
# Where the AgentQL WASM runtime and its JS glue live (consumed by the tooling).
AGENTQL_WEB := web/agentql-runtime
# Container manager for the local sandbox: podman (default) or docker.
CONTAINER ?= podman
IMAGE_PREFIX ?= localhost/unboxd-cloud
IMAGE_TAG ?= dev

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: build
build: ## Build all command binaries into ./bin
	@mkdir -p $(BIN)
	@for c in $(CMDS); do \
		echo "building $$c"; \
		$(GO) build -o $(BIN)/$$c ./cmd/$$c || exit 1; \
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

.PHONY: agentql-wasm
agentql-wasm: ## Build the AgentQL runtime as WASM for the TS tooling
	GOOS=js GOARCH=wasm $(GO) build -o $(AGENTQL_WEB)/agentql.wasm ./cmd/agentql-wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(AGENTQL_WEB)/wasm_exec.js
	@echo "AgentQL runtime built -> $(AGENTQL_WEB)/agentql.wasm"

.PHONY: e2e
e2e: ## End-to-end test: run the stack and exercise the full flow
	./scripts/e2e.sh

.PHONY: sanity
sanity: ## Post-deploy sanity checks (set *_URL env for a remote deployment)
	./scripts/post-deploy-sanity.sh

.PHONY: scan
scan: ## Scan Go packages for known vulnerabilities (govulncheck)
	@command -v govulncheck >/dev/null 2>&1 || $(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: scan-image
scan-image: ## Scan a built image for vulns (trivy). SVC=catalog
	$(CONTAINER) build --build-arg SERVICE=$(or $(SVC),catalog) -t scan/$(or $(SVC),catalog):local .
	trivy image --severity HIGH,CRITICAL --ignore-unfixed scan/$(or $(SVC),catalog):local

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN)

## --- Local sandbox (deployment-path testing with podman) ---

.PHONY: images
images: ## Build OCI images for all services ($(CONTAINER))
	@for svc in $(SERVICES); do \
		echo "building image $$svc"; \
		$(CONTAINER) build --build-arg SERVICE=$$svc \
			-t $(IMAGE_PREFIX)/$$svc:$(IMAGE_TAG) . || exit 1; \
	done

# One-click publish to any OCI registry (Docker Hub by default).
# Usage: make publish PUBLISH_REGISTRY=docker.io/youruser PUBLISH_TAG=0.1.0
PUBLISH_REGISTRY ?= docker.io/unboxd
PUBLISH_TAG ?= $(IMAGE_TAG)

.PHONY: publish
publish: ## Build, tag and push all images to $(PUBLISH_REGISTRY)
	@for svc in $(SERVICES); do \
		echo "publishing $$svc -> $(PUBLISH_REGISTRY)/$$svc:$(PUBLISH_TAG)"; \
		$(CONTAINER) build --build-arg SERVICE=$$svc \
			-t $(PUBLISH_REGISTRY)/$$svc:$(PUBLISH_TAG) . || exit 1; \
		$(CONTAINER) push $(PUBLISH_REGISTRY)/$$svc:$(PUBLISH_TAG) || exit 1; \
	done

.PHONY: sandbox-up
sandbox-up: images ## Run the full stack locally via 'podman play kube'
	$(CONTAINER) play kube deploy/sandbox/pod.yaml

.PHONY: sandbox-down
sandbox-down: ## Tear down the local sandbox
	$(CONTAINER) play kube --down deploy/sandbox/pod.yaml

.PHONY: sandbox-smoke
sandbox-smoke: ## Hit the running sandbox endpoints
	@echo "catalog:"   && curl -fs localhost:8083/v1/catalog | head -c 200; echo
	@echo "pricebook:"  && curl -fs localhost:8082/v1/pricebook | head -c 120; echo
	@echo "frameworks:" && curl -fs localhost:8084/v1/frameworks | head -c 200; echo
