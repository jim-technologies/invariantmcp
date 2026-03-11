package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type codexAdapter struct{}

type codexConfig struct {
	MCPServers map[string]codexServer `toml:"mcp_servers"`
}

type codexServer struct {
	Command string            `toml:"command"`
	Args    []string          `toml:"args,omitempty"`
	Env     map[string]string `toml:"env,omitempty"`
	EnvVars []string          `toml:"env_vars,omitempty"`
}

func (a *codexAdapter) ID() string { return "codex" }

func (a *codexAdapter) Match(path string) bool {
	return matchDirAndBase(path, ".codex", "config.toml")
}

func (a *codexAdapter) DefaultPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".codex", "config.toml"),
	}
}

func (a *codexAdapter) Import(data []byte) (*UniversalConfig, error) {
	var cc codexConfig
	if err := toml.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("parse codex config: %w", err)
	}
	cfg := &UniversalConfig{Servers: make(map[string]*ServerConfig)}
	for name, srv := range cc.MCPServers {
		cfg.Servers[name] = &ServerConfig{
			Transport:  "stdio",
			Command:    srv.Command,
			Args:       srv.Args,
			Env:        srv.Env,
			EnvInherit: srv.EnvVars,
		}
	}
	return cfg, nil
}

func (a *codexAdapter) Export(config *UniversalConfig) ([]byte, error) {
	cc := codexConfig{MCPServers: make(map[string]codexServer)}
	for name, srv := range config.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		cc.MCPServers[name] = codexServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			EnvVars: srv.EnvInherit,
		}
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cc); err != nil {
		return nil, fmt.Errorf("encode codex toml: %w", err)
	}
	return buf.Bytes(), nil
}

func (a *codexAdapter) Merge(existing []byte, source *UniversalConfig) ([]byte, int, error) {
	var root map[string]any
	if err := toml.Unmarshal(existing, &root); err != nil {
		return nil, 0, fmt.Errorf("parse codex config: %w", err)
	}
	servers := make(map[string]any)
	if raw, ok := root["mcp_servers"]; ok {
		existingServers, ok := raw.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("parse codex mcp_servers: unexpected type %T", raw)
		}
		for name, value := range existingServers {
			servers[name] = value
		}
	}
	for name, srv := range source.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		entry := map[string]any{
			"command": srv.Command,
		}
		if len(srv.Args) > 0 {
			entry["args"] = srv.Args
		}
		if len(srv.Env) > 0 {
			entry["env"] = srv.Env
		}
		if len(srv.EnvInherit) > 0 {
			entry["env_vars"] = srv.EnvInherit
		}
		servers[name] = entry
	}
	root["mcp_servers"] = servers
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(root); err != nil {
		return nil, 0, fmt.Errorf("encode codex toml: %w", err)
	}
	return buf.Bytes(), len(servers), nil
}
