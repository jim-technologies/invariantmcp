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
		{
			name: "claude code to codex",
			in:   "claudecode/.mcp.json",
			out:  ".codex/config.toml",
			want: "claudecode/.codex/config.toml",
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

func TestConvertMergesExistingTarget(t *testing.T) {
	svc := &ConfigService{}
	inPath := filepath.Join("testdata", "cursor", "mcp.json")
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, ".codex", "config.toml")

	existingBytes, err := os.ReadFile(filepath.Join("testdata", "merge", ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("read existing target fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	if err := os.WriteFile(outPath, existingBytes, 0o644); err != nil {
		t.Fatalf("write existing target fixture: %v", err)
	}

	resp, err := svc.Convert(context.Background(), &pb.ConvertRequest{
		In:  inPath,
		Out: outPath,
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	if resp.ServerCount != 3 {
		t.Fatalf("Convert() server count = %d, want 3", resp.ServerCount)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read merged output: %v", err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "merge", ".codex", "merged_config.toml"))
	if err != nil {
		t.Fatalf("read expected merged output: %v", err)
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(want)) {
		t.Fatalf("merged output mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestConvertPreservesUnrelatedTargetFields(t *testing.T) {
	svc := &ConfigService{}
	inPath := filepath.Join("testdata", "cursor", "mcp.json")
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "opencode.json")

	existingBytes, err := os.ReadFile(filepath.Join("testdata", "merge", "opencode.json"))
	if err != nil {
		t.Fatalf("read existing opencode fixture: %v", err)
	}
	if err := os.WriteFile(outPath, existingBytes, 0o644); err != nil {
		t.Fatalf("write existing opencode fixture: %v", err)
	}

	_, err = svc.Convert(context.Background(), &pb.ConvertRequest{
		In:  inPath,
		Out: outPath,
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read merged opencode output: %v", err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "merge", "merged_opencode.json"))
	if err != nil {
		t.Fatalf("read expected merged opencode output: %v", err)
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(want)) {
		t.Fatalf("merged opencode output mismatch\ngot:\n%s\nwant:\n%s", got, want)
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
		{"./.mcp.json", "claudecode"},
		{"/home/user/project/.mcp.json", "claudecode"},
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
