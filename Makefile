.PHONY: generate descriptor serve-mcp serve-cli test clean lint public-surface fmt build

generate:
	cd proto && buf generate
	cd proto && buf build -o ../descriptor.binpb

descriptor:
	cd proto && buf build -o ../descriptor.binpb

lint: public-surface
	cd proto && buf lint
	go vet ./...

# Guard the public surface: tracked content, tracked paths, and the commit
# messages a push would publish. Exceptions live in .public-surface-allow.
public-surface:
	scripts/public-surface-check
	scripts/public-surface-check-test

fmt:
	cd proto && buf format -w
	gofmt -w .

build:
	go build -o invariantmcp .

serve-mcp:
	go run .

serve-cli:
	go run . --cli

test:
	go test -v -count=1 ./...

clean:
	rm -f descriptor.binpb invariantmcp
	rm -rf mcpconfig/
