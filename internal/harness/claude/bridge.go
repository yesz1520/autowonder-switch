package claude

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type platformInput struct {
	Operation string          `json:"operation"`
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}
type commandInput struct {
	Args []string `json:"args"`
}
type questionInput struct {
	Message string         `json:"message"`
	Schema  map[string]any `json:"schema"`
}

func toolText(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

// The turn owns the listener and capability URL. Only scoped tool calls cross
// this boundary. The platform token stays in the Go executor.
func openBridge(ctx context.Context, q harness.Request, sandboxBinary string) (string, func(), error) {
	if q.Platform == nil && q.Ask == nil && !q.Readable && !q.Writable {
		return "", func() {}, nil
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "autowonder-switch", Version: "1"}, nil)
	if q.Platform != nil {
		mcp.AddTool(server, &mcp.Tool{Name: "platform", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"operation": map[string]any{"enum": []string{"list", "call"}}, "name": map[string]string{"type": "string"}, "arguments": map[string]any{"type": "object", "additionalProperties": true}}, "required": []string{"operation"}}, Description: "List available AutoWonder tools with operation=list, or call a listed tool with operation=call, name and arguments."}, func(ctx context.Context, _ *mcp.CallToolRequest, in platformInput) (*mcp.CallToolResult, any, error) {
			name := ""
			args := map[string]any{}
			switch in.Operation {
			case "list":
			case "call":
				if in.Name == "" {
					return nil, nil, errors.New("missing tool name")
				}
				name = in.Name
				if len(in.Arguments) > 0 {
					d := json.NewDecoder(strings.NewReader(string(in.Arguments)))
					d.UseNumber()
					if err := d.Decode(&args); err != nil {
						return nil, nil, err
					}
				}
			default:
				return nil, nil, errors.New("unknown operation")
			}
			value, err := q.Platform(ctx, name, args)
			if err != nil {
				return nil, nil, err
			}
			return toolText(value), nil, nil
		})
	}
	if q.Ask != nil {
		mcp.AddTool(server, &mcp.Tool{Name: "ask_user", Description: "Ask the platform user a question using a JSON schema for the answer."}, func(ctx context.Context, _ *mcp.CallToolRequest, in questionInput) (*mcp.CallToolResult, any, error) {
			a, err := q.Ask(ctx, in.Message, in.Schema)
			if err != nil {
				return nil, nil, err
			}
			b, err := json.Marshal(map[string]any{"action": a.Action, "content": a.Content})
			if a.Written != nil {
				a.Written(false)
			}
			return toolText(string(b)), nil, err
		})
	}
	if q.Writable || q.Readable {
		if sandboxBinary == "" {
			sandboxBinary = "codex"
		}
		mcp.AddTool(server, &mcp.Tool{Name: "workspace", Description: "Run a command in the executor workspace sandbox. Supply executable and arguments separately, e.g. [\"git\",\"status\"]."}, func(ctx context.Context, _ *mcp.CallToolRequest, in commandInput) (*mcp.CallToolResult, any, error) {
			if len(in.Args) == 0 {
				return nil, nil, errors.New("missing command")
			}
			mode := "read-only"
			if q.Writable {
				mode = "workspace-write"
			}
			args := []string{"sandbox", "-c", "sandbox_mode=\"" + mode + "\"", "-c", "sandbox_workspace_write.network_access=false", "--", "/usr/bin/env", "-i", "PATH=" + os.Getenv("PATH")}
			for k, v := range q.Environment {
				if k == "" || strings.ContainsAny(k, "=\x00") || strings.HasPrefix(k, "-") {
					return nil, nil, errors.New("invalid environment key")
				}
				args = append(args, k+"="+v)
			}
			args = append(args, in.Args...)
			cmd := exec.CommandContext(ctx, sandboxBinary, args...)
			cmd.Dir = q.Directory
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
			out := &limitedText{}
			cmd.Stdout = out
			cmd.Stderr = out
			err := cmd.Run()
			if err != nil {
				return toolText(out.text.String() + "\n" + err.Error()), nil, nil
			}
			return toolText(out.text.String()), nil, nil
		})
	}
	var secret [32]byte
	rand.Read(secret[:])
	path := "/" + hex.EncodeToString(secret[:])
	mux := http.NewServeMux()
	mux.Handle(path, mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	srv := &http.Server{Handler: mux, BaseContext: func(net.Listener) context.Context { return ctx }, ReadHeaderTimeout: 5 * time.Second}
	go srv.Serve(ln)
	cfg, err := json.Marshal(map[string]any{"mcpServers": map[string]any{"switch": map[string]string{"type": "http", "url": "http://" + ln.Addr().String() + path}}})
	return string(cfg), func() { srv.Close() }, err
}

type limitedText struct{ text strings.Builder }

func (w *limitedText) Write(p []byte) (int, error) {
	remaining := 65536 - w.text.Len()
	if remaining > 0 {
		w.text.Write(p[:min(remaining, len(p))])
	}
	return len(p), nil
}
