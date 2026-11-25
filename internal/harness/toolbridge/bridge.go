// Package toolbridge serves the tools scoped to one local harness turn.
package toolbridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/sandbox"
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
func Open(ctx context.Context, q harness.Request) (string, func(), error) {
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
		workspace := sandbox.Workspace{Directory: q.Directory, Writable: q.Writable, ReadOnlyPaths: q.ReadOnlyPaths}
		mcp.AddTool(server, &mcp.Tool{Name: "workspace", Description: "Run a command in the executor workspace sandbox. Supply executable and arguments separately, e.g. [\"git\",\"status\"]."}, func(ctx context.Context, _ *mcp.CallToolRequest, in commandInput) (*mcp.CallToolResult, any, error) {
			out := &limitedText{}
			err := workspace.Run(ctx, q.Environment, in.Args, out)
			if err != nil {
				result := toolText(out.text.String() + "\n" + err.Error())
				result.IsError = true
				return result, nil, nil
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
	return "http://" + ln.Addr().String() + path, func() { srv.Close() }, nil
}

type limitedText struct{ text strings.Builder }

func (w *limitedText) Write(p []byte) (int, error) {
	remaining := 65536 - w.text.Len()
	if remaining > 0 {
		w.text.Write(p[:min(remaining, len(p))])
	}
	return len(p), nil
}
