package taskhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToolCommands bridges native Codex stdin to the package hook contract. The
// shell wrapper fails closed when the listener is unavailable.
func (s *Set) ToolCommands(ctx context.Context) (map[string]string, func(), error) {
	if !s.Has("beforeTool") && !s.Has("afterTool") {
		return nil, func() {}, nil
	}
	curl, err := exec.LookPath("curl")
	if err != nil {
		return nil, nil, err
	}
	dir, err := os.MkdirTemp("", "aws-hook-")
	if err != nil {
		return nil, nil, err
	}
	socket := filepath.Join(dir, "h.sock")
	ln, err := net.Listen("unix", socket)
	if err != nil {
		os.RemoveAll(dir)
		return nil, nil, err
	}
	mux := http.NewServeMux()
	commands := map[string]string{}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	for event, native := range map[string]string{"beforeTool": "PreToolUse", "afterTool": "PostToolUse"} {
		if !s.Has(event) {
			continue
		}
		mux.HandleFunc("POST /"+event, func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			var v struct {
				ToolName   string          `json:"tool_name"`
				ToolInput  json.RawMessage `json:"tool_input"`
				ToolOutput json.RawMessage `json:"tool_response"`
				CallID     string          `json:"tool_use_id"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&v); err != nil || v.ToolName == "" {
				http.Error(w, "invalid tool context", 400)
				return
			}
			fields := map[string]any{"schemaVersion": "autowonder.toolHookContext.v1", "toolName": v.ToolName, "tool_name": v.ToolName, "toolInput": v.ToolInput, "tool_input": v.ToolInput}
			if v.CallID != "" {
				fields["callId"] = v.CallID
			}
			if len(v.ToolOutput) > 0 {
				fields["toolOutput"], fields["tool_output"] = v.ToolOutput, v.ToolOutput
			}
			if err := s.Run(r.Context(), event, "", fields); err != nil {
				http.Error(w, "package hook denied tool", 403)
				return
			}
			fmt.Fprint(w, "allow")
		})
		commands[native] = "response=$(" + quote(curl) + " --disable --silent --fail --noproxy '*' --unix-socket " + quote(socket) + " --data-binary @- http://localhost/" + event + ") && [ \"$response\" = allow ] || { echo 'AutoWonder package hook denied tool' >&2; exit 2; }"
	}
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	go server.Serve(ln)
	return commands, func() { server.Close(); os.RemoveAll(dir) }, nil
}
