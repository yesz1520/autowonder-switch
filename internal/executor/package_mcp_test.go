package executor

import (
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func echoMCP() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "fixture", Version: "1"}, nil)
	type input struct {
		Value string `json:"value"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "echo", Description: "Echo a value"}, func(_ context.Context, _ *mcp.CallToolRequest, in input) (*mcp.CallToolResult, input, error) {
		return nil, input{Value: "logged-response-" + in.Value}, nil
	})
	return s
}
func TestMCPHelperProcess(t *testing.T) {
	if os.Getenv("SWITCH_TEST_MCP_CHILD") != "true" {
		return
	}
	if os.Getenv("MCP_KEY") != "resolved-secret" || os.Getenv("AUTOWONDER_TOKEN") != "" {
		os.Exit(2)
	}
	if err := echoMCP().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestPackageMCPTransports(t *testing.T) {
	for _, mode := range []string{"http", "sse", "stdio"} {
		t.Run(mode, func(t *testing.T) {
			ctx, stop := context.WithTimeout(t.Context(), 90*time.Second)
			defer stop()
			srv := echoMCP()
			get := func(*http.Request) *mcp.Server { return srv }
			var endpoint http.Handler = mcp.NewStreamableHTTPHandler(get, nil)
			if mode == "sse" {
				endpoint = mcp.NewSSEHandler(get, nil)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Test-Key") != "resolved-secret" {
					w.WriteHeader(403)
					return
				}
				endpoint.ServeHTTP(w, r)
			}))
			defer server.Close()
			t.Setenv("AUTOWONDER_TOKEN", "never-inherit")
			cmd, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			desc := map[string]any{"name": "sample", "transport": mode, "url": server.URL, "headerSecretRefs": map[string]string{"X-Test-Key": "credential"}, "command": cmd, "args": []string{"-test.run=^TestMCPHelperProcess$"}, "env": map[string]string{"SWITCH_TEST_MCP_CHILD": "true"}, "envSecretRefs": map[string]string{"MCP_KEY": "credential"}, "required": true}
			doc, _ := json.Marshal(map[string]any{"mcpServers": []any{desc}})
			dir := t.TempDir()
			pkg := filepath.Join(dir, "test.zip")
			if err = os.WriteFile(pkg, archive(t, map[string]string{"workitem.md": "test", "skills.json": string(doc)}), 0600); err != nil {
				t.Fatal(err)
			}
			spec := harness.Request{Directory: dir}
			text, err := packageText(pkg, true, &spec)
			if err != nil || strings.Contains(text, "credential") || strings.Contains(text, "resolved-secret") {
				t.Fatal("unexpected capability metadata in prompt", err)
			}
			r := runtime{}
			closeTools, err := r.openPackageTools(ctx, pkg, map[string]string{"credential": "resolved-secret"}, &spec)
			if err != nil {
				t.Fatal(err)
			}
			defer closeTools()
			list, err := spec.Platform(ctx, "", nil)
			if err != nil || !strings.Contains(list, "sample/echo") {
				t.Fatal(list, err)
			}
			got, err := spec.Platform(ctx, "sample/echo", map[string]any{"value": "orchid753"})
			if err != nil || !strings.Contains(got, "orchid753") {
				t.Fatal(got, err)
			}
			if mode == "http" && os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
				spec.Prompt = "Invoke the autowonder dynamic tool with operation call, name sample/echo, and arguments value orchid753. Reply with the echoed value."
				reply, err := (codex.Backend{Binary: "codex"}).Run(ctx, spec, harness.Hooks{})
				if err != nil || !strings.Contains(reply.Text, "logged-response-orchid753") {
					t.Fatalf("reply=%q err=%v", reply.Text, err)
				}
			}
		})
	}
}

func TestPlatformMCPProbeRoundTrip(t *testing.T) {
	ctx, stop := context.WithTimeout(t.Context(), 10*time.Second)
	defer stop()
	srv := echoMCP()
	mux := http.NewServeMux()
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil))
	socket := make(chan *websocket.Conn, 1)
	mux.HandleFunc("/ws/executor", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		socket <- c
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	done := make(chan error, 1)
	dir := t.TempDir()
	go func() {
		done <- Run(ctx, aw.Config{ExecutorID: 1, BaseURL: server.URL, Token: "test"}, dir, &stub{}, nil)
	}()
	defer func() { stop(); <-done }()
	var c *websocket.Conn
	select {
	case c = <-socket:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	defer c.CloseNow()
	var heartbeat aw.Heartbeat
	if err := wsjson.Read(ctx, c, &heartbeat); err != nil || !slices.Contains(heartbeat.ProtocolFeatures, aw.FeatureActionPlanV1) {
		t.Fatal(heartbeat, err)
	}
	for _, success := range []bool{true, false} {
		path := "/mcp"
		if !success {
			path = "/missing"
		}
		err := wsjson.Write(ctx, c, map[string]any{"type": "MCP_CONNECTION_TEST", "testId": "ticket", "executorId": 1, "transport": "http", "url": server.URL + path, "timeoutSeconds": 3})
		if err != nil {
			t.Fatal(err)
		}
		var reply struct {
			Type, TestID, Message string
			Success               bool
			Tools                 []struct{ Name string }
		}
		if err = wsjson.Read(ctx, c, &reply); err != nil || reply.Type != "MCP_CONNECTION_TEST_RESULT" || reply.TestID != "ticket" || reply.Success != success {
			t.Fatal(reply, err)
		}
		if success && (len(reply.Tools) != 1 || reply.Tools[0].Name != "echo") {
			t.Fatal(reply)
		}
		if !success && reply.Message == "" {
			t.Fatal("missing diagnostic")
		}
	}
}

func TestPackageMCPMissingSecret(t *testing.T) {
	for _, required := range []bool{true, false} {
		dir := t.TempDir()
		pkg := filepath.Join(dir, "package.zip")
		doc, _ := json.Marshal(map[string]any{"mcpServers": []any{map[string]any{"name": "sample", "transport": "http", "url": "https://example.invalid", "headerSecretRefs": map[string]string{"Authorization": "private-ref"}, "required": required}}})
		if err := os.WriteFile(pkg, archive(t, map[string]string{"skills.json": string(doc)}), 0600); err != nil {
			t.Fatal(err)
		}
		spec := harness.Request{Directory: dir}
		r := runtime{}
		closeService, err := r.openPackageTools(t.Context(), pkg, nil, &spec)
		if required {
			if err == nil || !strings.Contains(err.Error(), "missing MCP secret reference") || strings.Contains(err.Error(), "private-ref") {
				t.Fatal(err)
			}
		} else {
			if err != nil || !strings.Contains(spec.Instructions, "sample") || spec.Platform != nil {
				t.Fatal(err, spec.Instructions)
			}
			closeService()
		}
	}
}
