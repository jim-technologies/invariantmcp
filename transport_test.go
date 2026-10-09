package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/jim-technologies/invariantmcp/mcpconfig/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestGeneratedRegistrationConvertsThroughConnect(t *testing.T) {
	server, err := configServer()
	if err != nil {
		t.Fatal(err)
	}
	host := httptest.NewServer(server.HTTPHandler())
	defer host.Close()
	source := filepath.Join("testdata", "cursor", "mcp.json")
	target := filepath.Join(t.TempDir(), "codex", ".codex", "config.toml")
	body, err := protojson.Marshal(&pb.ConvertRequest{In: source, Out: target})
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(host.URL+"/mcpconfig.v1.ConfigService/Convert", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("conversion returned %d: %s", response.StatusCode, data)
	}
	var result pb.ConvertResponse
	if err := protojson.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := (&codexAdapter{}).Import(written)
	if err != nil {
		t.Fatal(err)
	}
	if len(converted.Servers) == 0 || result.GetServerCount() != int32(len(converted.Servers)) || result.GetPath() != target {
		t.Fatalf("conversion lost servers or output identity: %v", &result)
	}
}
