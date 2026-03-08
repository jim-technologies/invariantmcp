package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Adapter defines the interface for bidirectional config translation.
type Adapter interface {
	// ID returns the unique identifier for this client, e.g. "cursor".
	ID() string
	// DefaultPaths returns candidate config file paths for the current OS.
	DefaultPaths() []string
	// Match reports whether the given file path belongs to this client.
	Match(path string) bool
	// Import parses raw config bytes into the universal IR.
	Import(data []byte) (*UniversalConfig, error)
	// Export translates the universal IR into the client's native format.
	Export(config *UniversalConfig) ([]byte, error)
}

var adapters []Adapter

func registerAdapter(a Adapter) {
	adapters = append(adapters, a)
}

func adapterForPath(path string) (Adapter, error) {
	for _, a := range adapters {
		if a.Match(path) {
			return a, nil
		}
	}
	return nil, fmt.Errorf("cannot detect client format from path %q", path)
}

// matchBasename checks if the file's base name matches.
func matchBasename(path, name string) bool {
	return filepath.Base(path) == name
}

// matchDirAndBase checks if the path ends with dir/base.
func matchDirAndBase(path, dir, base string) bool {
	return filepath.Base(path) == base &&
		strings.HasSuffix(filepath.Dir(path), dir)
}

func init() {
	registerAdapter(&cursorAdapter{})
	registerAdapter(&claudeAdapter{})
	registerAdapter(&codexAdapter{})
	registerAdapter(&opencodeAdapter{})
	registerAdapter(&clineAdapter{})
}
