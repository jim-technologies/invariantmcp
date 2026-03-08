package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type clineAdapter struct{}

type clineConfig struct {
	MCPServers map[string]clineServer `json:"mcpServers"`
}

type clineServer struct {
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	URL           string            `json:"url,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	TransportType string            `json:"transportType,omitempty"`
	Disabled      bool              `json:"disabled,omitempty"`
}

func (a *clineAdapter) ID() string { return "cline" }

func (a *clineAdapter) Match(path string) bool {
	return matchBasename(path, "cline_mcp_settings.json")
}

func (a *clineAdapter) DefaultPaths() []string {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings")
	if runtime.GOOS == "darwin" {
		base = filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings")
	}
	return []string{
		filepath.Join(base, "cline_mcp_settings.json"),
	}
}

func (a *clineAdapter) Import(data []byte) (*UniversalConfig, error) {
	var cc clineConfig
	if err := json.Unmarshal(data, &cc); err != nil {
		return nil, fmt.Errorf("parse cline config: %w", err)
	}
	cfg := &UniversalConfig{Servers: make(map[string]*ServerConfig)}
	for name, srv := range cc.MCPServers {
		if srv.Disabled {
			continue
		}
		sc := &ServerConfig{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			URL:     srv.URL,
			Headers: srv.Headers,
		}
		switch srv.TransportType {
		case "sse":
			sc.Transport = "http"
		default:
			sc.Transport = "stdio"
		}
		cfg.Servers[name] = sc
	}
	return cfg, nil
}

func (a *clineAdapter) Export(config *UniversalConfig) ([]byte, error) {
	cc := clineConfig{MCPServers: make(map[string]clineServer)}
	for name, srv := range config.Servers {
		cs := clineServer{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			Disabled: srv.Disabled,
		}
		if srv.Transport == "http" {
			cs.TransportType = "sse"
			cs.URL = srv.URL
			cs.Headers = srv.Headers
		} else {
			cs.TransportType = "stdio"
		}
		cc.MCPServers[name] = cs
	}
	return json.MarshalIndent(cc, "", "  ")
}
