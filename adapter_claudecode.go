package main

import (
	"encoding/json"
	"fmt"
)

type claudeCodeAdapter struct{}

type claudeCodeConfig struct {
	MCPServers map[string]claudeCodeServer `json:"mcpServers"`
}

type claudeCodeServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (a *claudeCodeAdapter) ID() string { return "claudecode" }

func (a *claudeCodeAdapter) Match(path string) bool {
	return matchBasename(path, ".mcp.json")
}

func (a *claudeCodeAdapter) DefaultPaths() []string {
	return []string{".mcp.json"}
}

func (a *claudeCodeAdapter) Import(data []byte) (*UniversalConfig, error) {
	var cc claudeCodeConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("parse claude code config: %w", err)
	}
	cfg := &UniversalConfig{Servers: make(map[string]*ServerConfig)}
	for name, srv := range cc.MCPServers {
		cfg.Servers[name] = &ServerConfig{
			Transport: "stdio",
			Command:   srv.Command,
			Args:      srv.Args,
			Env:       srv.Env,
		}
	}
	return cfg, nil
}

func (a *claudeCodeAdapter) Export(config *UniversalConfig) ([]byte, error) {
	cc := claudeCodeConfig{MCPServers: make(map[string]claudeCodeServer)}
	for name, srv := range config.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		cc.MCPServers[name] = claudeCodeServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		}
	}
	return json.MarshalIndent(cc, "", "  ")
}

func (a *claudeCodeAdapter) Merge(existing []byte, source *UniversalConfig) ([]byte, int, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(existing, &root); err != nil {
		return nil, 0, fmt.Errorf("parse claude code config: %w", err)
	}
	servers := make(map[string]json.RawMessage)
	if raw, ok := root["mcpServers"]; ok {
		if err := json.Unmarshal(raw, &servers); err != nil {
			return nil, 0, fmt.Errorf("parse claude code mcpServers: %w", err)
		}
	}
	for name, srv := range source.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		entry, err := json.Marshal(claudeCodeServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("encode claude code server %q: %w", name, err)
		}
		servers[name] = entry
	}
	rawServers, err := json.Marshal(servers)
	if err != nil {
		return nil, 0, fmt.Errorf("encode claude code mcpServers: %w", err)
	}
	root["mcpServers"] = rawServers
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("encode claude code config: %w", err)
	}
	return out, len(servers), nil
}
