package main

// UniversalConfig is the intermediate representation holding all managed MCP servers.
type UniversalConfig struct {
	Servers map[string]*ServerConfig `json:"servers"`
}

// ServerConfig defines the execution parameters for a single MCP server.
type ServerConfig struct {
	Transport  string            `json:"transport"`
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	URL        string            `json:"url,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	EnvInherit []string          `json:"env_inherit,omitempty"`
	Disabled   bool              `json:"disabled,omitempty"`
}
