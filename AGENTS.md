# AGENTS.md

Guide for AI agents and human contributors working on this codebase.

## What this project does

Translates MCP config files between AI clients. Takes an input file path and output file path, detects formats from the paths, converts through a universal IR. All fields are strongly typed — no `map[string]any`.

## Project structure

```
main.go              Entry point, Invariant Protocol projections (CLI / MCP / HTTP)
service.go           ConfigService.Convert RPC handler
ir.go                Universal intermediate representation types (fully typed)
adapter.go           Adapter interface, path-based detection, init()
adapter_cursor.go    Cursor adapter (JSON, matches mcp.json)
adapter_claude.go    Claude Desktop adapter (JSON, matches claude_desktop_config.json)
adapter_codex.go     Codex adapter (TOML, matches .codex/config.toml)
adapter_opencode.go  OpenCode adapter (JSONC, matches opencode.json)
adapter_cline.go     Cline adapter (JSON, matches cline_mcp_settings.json)
service_test.go      End-to-end conversion tests + path detection tests
proto/               Protobuf service definition (single Convert RPC)
mcpconfig/           Generated Go proto types (do not edit)
descriptor.binpb     Generated descriptor (do not edit)
testdata/            Test fixtures using real config filenames
  cursor/mcp.json
  claude/claude_desktop_config.json
  codex/.codex/config.toml
  opencode/opencode.json
  opencode/opencode_output.json
  cline/cline_mcp_settings.json
```

## How to add a new adapter

1. Create `adapter_<client>.go` in the root directory.
2. Implement the `Adapter` interface:

```go
type Adapter interface {
    ID() string
    DefaultPaths() []string
    Match(path string) bool
    Import(data []byte) (*UniversalConfig, error)
    Export(config *UniversalConfig) ([]byte, error)
}
```

- `Match` returns true if the file path belongs to this client (use `matchBasename` or `matchDirAndBase` helpers from `adapter.go`).
- `Import` parses the client's native format into `*UniversalConfig`.
- `Export` serializes `*UniversalConfig` into the client's native format.

3. Register it in `adapter.go` `init()`.
4. Add test fixtures in `testdata/<client>/` using the real config filename.
5. Add test cases to `service_test.go` (both conversion and path detection).

No other files need to change.

## The IR

All fields are concrete types. No `map[string]any` or `interface{}`.

```go
type ServerConfig struct {
    Transport  string
    Command    string
    Args       []string
    URL        string
    Headers    map[string]string
    Env        map[string]string
    EnvInherit []string
    Disabled   bool
}
```

If a new client has a field that doesn't fit, add it as a typed field to `ServerConfig` rather than using a catch-all map. Every field across every supported client maps to a concrete type.

## Rules

1. Everything is `package main`. No sub-packages except generated proto code.
2. Each adapter is one file. Keep it self-contained.
3. The IR types in `ir.go` are the source of truth.
4. Do not modify generated files in `mcpconfig/` or `descriptor.binpb`. Run `make generate` instead.
5. No `map[string]any` or `interface{}` in the IR. Add typed fields instead.
6. Test fixtures must use real config filenames so `Match` works in tests.

## Build commands

```
make generate    # Regenerate proto stubs + descriptor
make build       # Build binary
make test        # Run tests
make lint        # Public-surface guard, then lint proto + Go
make serve-mcp   # Run as MCP server
make serve-cli   # Run as CLI
```

`make lint` runs `make public-surface` first, and a finding there stops the
build. `scripts/public-surface-check` scans the content of every tracked file,
every tracked path, and the commit messages a push would publish, and fails on
private repository names, internal infrastructure, first-party codenames,
cluster shapes, credential shapes, secret stores and private remotes;
`scripts/public-surface-check-test` then proves the guard still works. The
script is identical in every public repository in the organisation, so never
edit it — justified exceptions go in `.public-surface-allow` at the repo root,
one line each as `category | path-glob | reason | pattern`.

## Do not

- Add sub-packages or internal directories. Keep it flat.
- Add dependencies unless absolutely necessary (e.g. a new config format parser).
- Over-engineer the adapter. 50-80 lines is the target.
- Add features to the proto service without a concrete use case.
- Modify `service.go` when adding a new adapter. Path detection handles dispatch.
- Use `map[string]any` or `interface{}` anywhere in the IR.
