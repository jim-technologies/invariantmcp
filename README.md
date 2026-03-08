# invariantmcp

Translate MCP config files between AI clients. Give it an input file and an output file — it detects the formats from the paths and handles the conversion.

Single Go binary. Built on [Invariant Protocol](https://github.com/jim-technologies/invariantprotocol).

## The problem

Every AI client has its own MCP config format:

| Client | Format | File |
|--------|--------|------|
| Cursor | JSON | `mcp.json` |
| Claude Desktop | JSON | `claude_desktop_config.json` |
| Codex | TOML | `.codex/config.toml` |
| OpenCode | JSONC | `opencode.json` |
| Cline | JSON | `cline_mcp_settings.json` |

Same servers, different files, different schemas. Add an MCP server and you're copy-pasting across all of them.

## How it works

One RPC: **Convert**. Two fields: **in** and **out**.

The client format is inferred from the file path. `mcp.json` means Cursor, `claude_desktop_config.json` means Claude, `.codex/config.toml` means Codex, `opencode.json` means OpenCode, `cline_mcp_settings.json` means Cline.

```
source file ── parse ── IR (Go struct) ── serialize ── target file
```

> **Note:** Convert overwrites the target file.

## Install

```bash
go install github.com/jim-technologies/invariantmcp@latest
```

Or build from source:

```bash
git clone https://github.com/jim-technologies/invariantmcp.git
cd invariantmcp
make build
```

## Usage

### CLI

```bash
# Cursor → Codex
invariantmcp --cli ConfigService Convert -r '{"in":"~/.cursor/mcp.json","out":"~/.codex/config.toml"}'

# Claude Desktop → Cursor
invariantmcp --cli ConfigService Convert -r '{"in":"~/Library/Application Support/Claude/claude_desktop_config.json","out":"~/.cursor/mcp.json"}'

# Project-level
invariantmcp --cli ConfigService Convert -r '{"in":".cursor/mcp.json","out":".claude/claude_desktop_config.json"}'

# See help
invariantmcp --cli --help
```

### MCP tool

```bash
invariantmcp
```

Add to your AI client's MCP config:

```json
{
  "mcpServers": {
    "invariantmcp": {
      "command": "invariantmcp"
    }
  }
}
```

Your AI agent can then call `ConfigService.Convert` to translate configs between clients.

### HTTP

```bash
invariantmcp --http 8080
```

```bash
curl -X POST http://localhost:8080/mcpconfig.v1.ConfigService/Convert \
  -d '{"in":"~/.cursor/mcp.json","out":"~/.codex/config.toml"}'
```

All three interfaces come from the same protobuf definition via Invariant Protocol.

## Example

Cursor config (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxx"
      }
    }
  }
}
```

```bash
invariantmcp --cli ConfigService Convert -r '{"in":"~/.cursor/mcp.json","out":"~/.codex/config.toml"}'
```

Output (`~/.codex/config.toml`):

```toml
[mcp_servers.github]
  command = "npx"
  args = ["-y", "@modelcontextprotocol/server-github"]
  [mcp_servers.github.env]
    GITHUB_PERSONAL_ACCESS_TOKEN = "ghp_xxx"
```

## Supported clients

| Client | Format | Detected by | Default path | Notes |
|--------|--------|-------------|--------------|-------|
| Cursor | JSON | `mcp.json` | `~/.cursor/mcp.json` | |
| Claude Desktop | JSON | `claude_desktop_config.json` | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS), `~/.config/claude/claude_desktop_config.json` (Linux) | |
| Codex | TOML | `.codex/config.toml` | `~/.codex/config.toml` | Preserves `env_vars` pass-through |
| OpenCode | JSONC | `opencode.json` | `~/.config/opencode/opencode.json` | Strips comments on import, skips disabled servers |
| Cline | JSON | `cline_mcp_settings.json` | `~/.config/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` | Skips disabled servers, maps SSE transport |

### Not yet supported

These clients have MCP config support but don't have adapters yet. PRs welcome.

| Client | Format | File | Notes |
|--------|--------|------|-------|
| Windsurf | JSON | `~/.codeium/windsurf/mcp_config.json` | Same `mcpServers` schema as Cursor |
| Roo Code | JSON | `.roo/mcp.json` (project), `mcp_settings.json` (global) | Cline fork, similar schema |
| Claude Code | JSON | `.mcp.json` | Same `mcpServers` schema, supports `${VAR}` expansion |
| VS Code Copilot | JSON | `.vscode/mcp.json` | Uses `"servers"` key instead of `"mcpServers"` |
| Zed | JSON | `~/.config/zed/settings.json` | Uses `"context_servers"` key, requires `"source": "custom"` |
| Continue | YAML | `~/.continue/config.yaml` | YAML format, uses array instead of map |

## Adding a new adapter

Each adapter is one file (~50-80 lines). Here's the full process:

### 1. Create `adapter_<client>.go`

Implement the `Adapter` interface:

```go
type Adapter interface {
    ID() string              // e.g. "windsurf"
    DefaultPaths() []string  // where the config lives on disk
    Match(path string) bool  // detect this client from a file path
    Import(data []byte) (*UniversalConfig, error)  // native format → IR
    Export(config *UniversalConfig) ([]byte, error) // IR → native format
}
```

Example — a Windsurf adapter would look like:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

type windsurfAdapter struct{}

type windsurfConfig struct {
    MCPServers map[string]windsurfServer `json:"mcpServers"`
}

type windsurfServer struct {
    Command string            `json:"command"`
    Args    []string          `json:"args,omitempty"`
    Env     map[string]string `json:"env,omitempty"`
}

func (a *windsurfAdapter) ID() string { return "windsurf" }

func (a *windsurfAdapter) Match(path string) bool {
    return matchBasename(path, "mcp_config.json")
}

func (a *windsurfAdapter) DefaultPaths() []string {
    home, _ := os.UserHomeDir()
    return []string{
        filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"),
    }
}

func (a *windsurfAdapter) Import(data []byte) (*UniversalConfig, error) {
    var wc windsurfConfig
    if err := json.Unmarshal(data, &wc); err != nil {
        return nil, fmt.Errorf("parse windsurf config: %w", err)
    }
    cfg := &UniversalConfig{Servers: make(map[string]*ServerConfig)}
    for name, srv := range wc.MCPServers {
        cfg.Servers[name] = &ServerConfig{
            Transport: "stdio",
            Command:   srv.Command,
            Args:      srv.Args,
            Env:       srv.Env,
        }
    }
    return cfg, nil
}

func (a *windsurfAdapter) Export(config *UniversalConfig) ([]byte, error) {
    wc := windsurfConfig{MCPServers: make(map[string]windsurfServer)}
    for name, srv := range config.Servers {
        if srv.Transport != "stdio" {
            continue
        }
        wc.MCPServers[name] = windsurfServer{
            Command: srv.Command,
            Args:    srv.Args,
            Env:     srv.Env,
        }
    }
    return json.MarshalIndent(wc, "", "  ")
}
```

### 2. Register it

In `adapter.go`, add to `init()`:

```go
registerAdapter(&windsurfAdapter{})
```

### 3. Add test fixtures

Create `testdata/windsurf/mcp_config.json` with a sample config using the real filename.

### 4. Add test cases

In `service_test.go`, add entries to both `TestConvert` (end-to-end conversion) and `TestAdapterForPath` (path detection).

That's it. `service.go` and the proto definition don't change.

### Path detection helpers

Two helpers in `adapter.go` for implementing `Match`:

- `matchBasename(path, "mcp_config.json")` — matches any path ending in that filename
- `matchDirAndBase(path, ".codex", "config.toml")` — matches when the parent dir and filename both match (use this when the filename alone is too generic)

### The IR

All fields are strongly typed — no `map[string]any`:

```go
type ServerConfig struct {
    Transport  string            // "stdio" or "http"
    Command    string            // executable (stdio)
    Args       []string          // command arguments (stdio)
    URL        string            // server URL (http/sse)
    Headers    map[string]string // HTTP headers (http/sse)
    Env        map[string]string // environment variables
    EnvInherit []string          // env var names to pass through (Codex)
    Disabled   bool              // server is disabled
}
```

If the new client has a field that doesn't exist in the IR, add it as a typed field to `ServerConfig` in `ir.go`. Do not use `map[string]any`.

## License

Apache 2.0
