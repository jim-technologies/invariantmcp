package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type opencodeAdapter struct{}

// opencodeConfig represents the relevant subset of opencode.json.
type opencodeConfig struct {
	MCP map[string]opencodeServer `json:"mcp"`
}

type opencodeServer struct {
	Type        string            `json:"type,omitempty"`
	Command     []string          `json:"command,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

func (a *opencodeAdapter) ID() string { return "opencode" }

func (a *opencodeAdapter) Match(path string) bool {
	return matchBasename(path, "opencode.json")
}

func (a *opencodeAdapter) DefaultPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "opencode", "opencode.json"),
	}
}

func (a *opencodeAdapter) Import(data []byte) (*UniversalConfig, error) {
	cleaned := stripJSONCComments(data)
	var oc opencodeConfig
	if err := json.Unmarshal(cleaned, &oc); err != nil {
		return nil, fmt.Errorf("parse opencode config: %w", err)
	}
	cfg := &UniversalConfig{Servers: make(map[string]*ServerConfig)}
	for name, srv := range oc.MCP {
		if srv.Enabled != nil && !*srv.Enabled {
			continue
		}
		sc := &ServerConfig{
			Env: srv.Environment,
		}
		if srv.Type == "remote" || srv.URL != "" {
			sc.Transport = "http"
			sc.URL = srv.URL
			sc.Headers = srv.Headers
		} else {
			sc.Transport = "stdio"
			if len(srv.Command) > 0 {
				sc.Command = srv.Command[0]
				sc.Args = srv.Command[1:]
			}
		}
		cfg.Servers[name] = sc
	}
	return cfg, nil
}

func (a *opencodeAdapter) Export(config *UniversalConfig) ([]byte, error) {
	oc := opencodeConfig{MCP: make(map[string]opencodeServer)}
	enabled := true
	for name, srv := range config.Servers {
		os := opencodeServer{
			Environment: srv.Env,
			Enabled:     &enabled,
		}
		if srv.Transport == "http" {
			os.Type = "remote"
			os.URL = srv.URL
			os.Headers = srv.Headers
		} else {
			os.Type = "local"
			os.Command = append([]string{srv.Command}, srv.Args...)
		}
		oc.MCP[name] = os
	}
	return json.MarshalIndent(oc, "", "  ")
}

func (a *opencodeAdapter) Merge(existing []byte, source *UniversalConfig) ([]byte, int, error) {
	cleaned := stripJSONCComments(existing)
	var root map[string]json.RawMessage
	if err := json.Unmarshal(cleaned, &root); err != nil {
		return nil, 0, fmt.Errorf("parse opencode config: %w", err)
	}
	servers := make(map[string]json.RawMessage)
	if raw, ok := root["mcp"]; ok {
		if err := json.Unmarshal(raw, &servers); err != nil {
			return nil, 0, fmt.Errorf("parse opencode mcp: %w", err)
		}
	}
	enabled := true
	for name, srv := range source.Servers {
		entry := opencodeServer{
			Environment: srv.Env,
			Enabled:     &enabled,
		}
		if srv.Transport == "http" {
			entry.Type = "remote"
			entry.URL = srv.URL
			entry.Headers = srv.Headers
		} else {
			entry.Type = "local"
			entry.Command = append([]string{srv.Command}, srv.Args...)
		}
		rawEntry, err := json.Marshal(entry)
		if err != nil {
			return nil, 0, fmt.Errorf("encode opencode server %q: %w", name, err)
		}
		servers[name] = rawEntry
	}
	rawServers, err := json.Marshal(servers)
	if err != nil {
		return nil, 0, fmt.Errorf("encode opencode mcp: %w", err)
	}
	root["mcp"] = rawServers
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, 0, fmt.Errorf("encode opencode config: %w", err)
	}
	return out, len(servers), nil
}

// stripJSONCComments removes // line comments and /* block comments */ from JSONC.
func stripJSONCComments(data []byte) []byte {
	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		// Inside a string literal — pass through unchanged
		if data[i] == '"' {
			out = append(out, data[i])
			i++
			for i < len(data) && data[i] != '"' {
				if data[i] == '\\' && i+1 < len(data) {
					out = append(out, data[i], data[i+1])
					i += 2
					continue
				}
				out = append(out, data[i])
				i++
			}
			if i < len(data) {
				out = append(out, data[i])
				i++
			}
			continue
		}
		// Line comment
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '/' {
			i += 2
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		}
		// Block comment
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '*' {
			i += 2
			for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			if i+1 < len(data) {
				i += 2
			}
			continue
		}
		out = append(out, data[i])
		i++
	}
	return out
}
