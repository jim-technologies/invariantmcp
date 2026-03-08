package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/jim-technologies/invariantmcp/mcpconfig/v1"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name string
		in   string // relative to testdata
		out  string // basename for temp output
		want string // relative to testdata, expected output to compare against
	}{
		{
			name: "cursor to codex",
			in:   "cursor/mcp.json",
			out:  ".codex/config.toml",
			want: "codex/.codex/config.toml",
		},
		{
			name: "cursor to claude",
			in:   "cursor/mcp.json",
			out:  "claude_desktop_config.json",
			want: "claude/claude_desktop_config.json",
		},
		{
			name: "opencode to cursor",
			in:   "opencode/opencode.json",
			out:  "mcp.json",
			want: "",
		},
		{
			name: "cline to cursor",
			in:   "cline/cline_mcp_settings.json",
			out:  "mcp.json",
			want: "",
		},
		{
			name: "cursor to cline",
			in:   "cursor/mcp.json",
			out:  "cline_mcp_settings.json",
			want: "",
		},
	}

	svc := &ConfigService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inPath := filepath.Join("testdata", tt.in)
			outDir := t.TempDir()
			outPath := filepath.Join(outDir, tt.out)

			resp, err := svc.Convert(context.Background(), &pb.ConvertRequest{
				In:  inPath,
				Out: outPath,
			})
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}

			if resp.ServerCount == 0 {
				t.Fatal("expected at least 1 server")
			}

			got, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}

			if tt.want != "" {
				wantBytes, err := os.ReadFile(filepath.Join("testdata", tt.want))
				if err != nil {
					t.Fatalf("read expected: %v", err)
				}
				if strings.TrimSpace(string(got)) != strings.TrimSpace(string(wantBytes)) {
					t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, wantBytes)
				}
			}
		})
	}
}

func TestConvertUnknownPath(t *testing.T) {
	svc := &ConfigService{}
	_, err := svc.Convert(context.Background(), &pb.ConvertRequest{
		In:  "some/random/file.txt",
		Out: "another/random/file.txt",
	})
	if err == nil {
		t.Fatal("expected error for unknown path")
	}
}

func TestAdapterForPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"~/.cursor/mcp.json", "cursor"},
		{".cursor/mcp.json", "cursor"},
		{"/home/user/project/mcp.json", "cursor"},
		{"~/Library/Application Support/Claude/claude_desktop_config.json", "claude"},
		{"/home/user/.config/claude/claude_desktop_config.json", "claude"},
		{"/home/user/.codex/config.toml", "codex"},
		{"/home/user/.config/opencode/opencode.json", "opencode"},
		{"/home/user/.config/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json", "cline"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			a, err := adapterForPath(tt.path)
			if err != nil {
				t.Fatalf("adapterForPath(%q) error: %v", tt.path, err)
			}
			if a.ID() != tt.want {
				t.Errorf("adapterForPath(%q) = %q, want %q", tt.path, a.ID(), tt.want)
			}
		})
	}
}
