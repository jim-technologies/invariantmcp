package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/jim-technologies/invariantmcp/mcpconfig/v1"
	invariant "github.com/jim-technologies/invariantprotocol/go"
)

//go:embed descriptor.binpb
var descriptorBytes []byte

func configServer() (*invariant.Server, error) {
	server, err := invariant.ServerFromBytes(descriptorBytes)
	if err != nil {
		return nil, err
	}
	pb.RegisterConfigServiceServer(server, &ConfigService{})
	return server, nil
}

func main() {
	server, err := configServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--cli" {
		os.Args = append([]string{os.Args[0]}, args[1:]...)
		if err := server.ServeProjections(ctx, invariant.CLI()); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 && args[0] == "--http" {
		port := 8080
		if len(args) > 1 {
			if p, err := strconv.Atoi(args[1]); err == nil {
				port = p
			}
		}
		if err := server.ServeProjections(ctx, invariant.HTTP(port)); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := server.ServeProjections(ctx, invariant.MCP()); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
