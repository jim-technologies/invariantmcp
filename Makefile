.PHONY: validate generate descriptor serve-mcp serve-cli test clean lint public-surface fmt build help help-all

generate: ## Regenerate protobuf code and descriptor.binpb from proto/
	cd proto && buf generate
	cd proto && buf build -o ../descriptor.binpb

descriptor: ## Rebuild descriptor.binpb only (no code generation)
	cd proto && buf build -o ../descriptor.binpb

# The gate. `make validate` is the one gate verb in every public repository in
# this organisation; here it routes to `lint` and `test`, this repo's full gate.
validate: lint test ## The full offline gate (lint + test)

lint: public-surface ## Public-surface guard, buf lint, go vet
	cd proto && buf lint
	go vet ./...

# Guard the public surface: tracked content, tracked paths, and the commit
# messages a push would publish. Exceptions live in .public-surface-allow.
public-surface: ## Public-surface guard and its self-test
	scripts/public-surface-check
	scripts/public-surface-check-test

fmt: ## Autofix formatting (buf format + gofmt)
	cd proto && buf format -w
	gofmt -w .

build: ## Build the invariantmcp binary
	go build -o invariantmcp .

serve-mcp: ## Run the MCP server (stdio)
	go run .

serve-cli: ## Run the CLI entrypoint (go run . --cli)
	go run . --cli

test: ## Unit tests (go test, uncached)
	go test -v -count=1 ./...

clean: ## Remove built binary, descriptor.binpb, and mcpconfig/
	rm -f descriptor.binpb invariantmcp
	rm -rf mcpconfig/

help: ## One-screen help (make help-all for every target)
	@echo "Daily:"
	@echo "  make fmt        autofix formatting (buf format + gofmt)"
	@echo "  make test       unit tests"
	@echo "  make validate   the full offline gate (lint + test)"
	@echo "  make generate   regenerate protobuf code + descriptor.binpb"
	@echo "  make serve-mcp  run the MCP server (stdio)"
	@echo ""
	@echo "Everything else: make help-all"

help-all: ## Every target with its description
	@grep -hE '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | sed -E 's/:.*## /\t/'
