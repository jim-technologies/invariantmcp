package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type cursorAdapter struct{}

type cursorConfig struct {
	MCPServers map[string]cursorServer `json:"mcpServers"`
}

type cursorServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (a *cursorAdapter) ID() string { return "cursor" }

func (a *cursorAdapter) Match(path string) bool {
	return matchBasename(path, "mcp.json")
}

func (a *cursorAdapter) DefaultPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".cursor", "mcp.json"),
	}
}

func (a *cursorAdapter) Import(data []byte) (*UniversalConfig, error) {
	var cc cursorConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("parse cursor config: %w", err)
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

func (a *cursorAdapter) Export(config *UniversalConfig) ([]byte, error) {
	cc := cursorConfig{MCPServers: make(map[string]cursorServer)}
	for name, srv := range config.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		cc.MCPServers[name] = cursorServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		}
	}
	return json.MarshalIndent(cc, "", "  ")
}
