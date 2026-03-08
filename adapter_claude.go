package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type claudeAdapter struct{}

type claudeConfig struct {
	MCPServers map[string]claudeServer `json:"mcpServers"`
}

type claudeServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (a *claudeAdapter) ID() string { return "claude" }

func (a *claudeAdapter) Match(path string) bool {
	return matchBasename(path, "claude_desktop_config.json")
}

func (a *claudeAdapter) DefaultPaths() []string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".config", "claude", "claude_desktop_config.json"),
	}
	if runtime.GOOS == "darwin" {
		paths = append([]string{
			filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
		}, paths...)
	}
	return paths
}

func (a *claudeAdapter) Import(data []byte) (*UniversalConfig, error) {
	var cc claudeConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("parse claude config: %w", err)
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

func (a *claudeAdapter) Export(config *UniversalConfig) ([]byte, error) {
	cc := claudeConfig{MCPServers: make(map[string]claudeServer)}
	for name, srv := range config.Servers {
		if srv.Transport != "stdio" {
			continue
		}
		cc.MCPServers[name] = claudeServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		}
	}
	return json.MarshalIndent(cc, "", "  ")
}
