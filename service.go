package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/jim-technologies/invariantmcp/mcpconfig/v1"
)

// ConfigService implements the mcpconfig.v1.ConfigService RPCs.
type ConfigService struct{}

func (s *ConfigService) Convert(_ context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	inPath := expandHome(req.In)
	outPath := expandHome(req.Out)

	from, err := adapterForPath(inPath)
	if err != nil {
		return nil, fmt.Errorf("source: %w", err)
	}
	to, err := adapterForPath(outPath)
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}

	data, err := os.ReadFile(inPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", inPath, err)
	}
	cfg, err := from.Import(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", from.ID(), err)
	}

	out, count, err := exportOrMergeTarget(outPath, to, cfg)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, fmt.Errorf("create dir for %s: %w", outPath, err)
	}
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return nil, fmt.Errorf("write %s: %w", outPath, err)
	}

	return &pb.ConvertResponse{
		ServerCount: int32(count),
		Path:        outPath,
		Message:     fmt.Sprintf("wrote %d merged servers from %s to %s (%s)", count, from.ID(), to.ID(), outPath),
	}, nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func exportOrMergeTarget(path string, adapter Adapter, source *UniversalConfig) ([]byte, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			out, err := adapter.Export(source)
			if err != nil {
				return nil, 0, fmt.Errorf("export %s: %w", adapter.ID(), err)
			}
			return out, len(source.Servers), nil
		}
		return nil, 0, fmt.Errorf("read %s: %w", path, err)
	}
	out, count, err := adapter.Merge(data, source)
	if err != nil {
		return nil, 0, fmt.Errorf("merge existing %s target: %w", adapter.ID(), err)
	}
	return out, count, nil
}
