package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/claude"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise the installed CLI, including its actual SSE parser, MCP client and
// transcript store, against a deterministic loopback Messages API.
func claudeFixture(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Messages json.RawMessage
			Model    string
			Tools    []struct{ Name string }
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		text := string(req.Messages)
		var tool string
		var input string
		if !strings.Contains(text, "tool_result") {
			for _, v := range req.Tools {
				switch {
				case strings.Contains(text, "Request a handoff") && v.Name == "mcp__switch__platform":
					tool = v.Name
					input = `{"operation":"call","name":"executor.request_handoff","arguments":{"to":"reviewer","toType":"AGENT"}}`
				case strings.Contains(text, "workspace-test") && v.Name == "mcp__switch__workspace":
					tool = v.Name
					input = `{"args":["sh","-c","echo verified > artifact.txt"]}`
				}
			}
		}
		content := map[string]any{"type": "text", "text": ""}
		answer := "violet"
		if strings.Contains(text, "confirm-prior-turn") {
			answer = "lost-history"
			if strings.Contains(text, "first-turn-marker") {
				answer = "preserved-history"
			}
		}
		delta := map[string]any{"type": "text_delta", "text": answer}
		reason := "end_turn"
		if tool != "" {
			content = map[string]any{"type": "tool_use", "id": "toolu_handoff", "name": tool, "input": map[string]any{}}
			delta = map[string]any{"type": "input_json_delta", "partial_json": input}
			reason = "tool_use"
		}
		w.Header().Set("Content-Type", "text/event-stream")
		emit := func(kind string, fields map[string]any) {
			fields["type"] = kind
			b, _ := json.Marshal(fields)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, b)
		}
		emit("message_start", map[string]any{"message": map[string]any{"id": "msg_test", "role": "assistant", "type": "message", "model": req.Model, "content": []any{}, "usage": map[string]int{"input_tokens": 20, "output_tokens": 0}}})
		emit("content_block_start", map[string]any{"index": 0, "content_block": content})
		emit("content_block_delta", map[string]any{"index": 0, "delta": delta})
		emit("content_block_stop", map[string]any{"index": 0})
		emit("message_delta", map[string]any{"delta": map[string]string{"stop_reason": reason}, "usage": map[string]int{"output_tokens": 5}})
		emit("message_stop", map[string]any{})
	}))
	t.Cleanup(srv.Close)
	for k, v := range map[string]string{"CLAUDE_CONFIG_DIR": t.TempDir(), "ANTHROPIC_BASE_URL": srv.URL, "ANTHROPIC_API_KEY": "fixture", "ANTHROPIC_AUTH_TOKEN": "", "CLAUDE_CODE_OAUTH_TOKEN": "", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1", "ANTHROPIC_MODEL": "claude-sonnet-4-6"} {
		t.Setenv(k, v)
	}
}

func TestClaudeNativeWorkspace(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CLAUDE") != "1" {
		t.Skip("requires installed Claude Code and sandbox CLI")
	}
	claudeFixture(t)
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	pre := filepath.Join(dir, "pre-ran")
	post := filepath.Join(dir, "post-ran")
	_, err := (claude.Backend{}).Run(ctx, harness.Request{Directory: dir, Writable: true, Prompt: "workspace-test", ToolHookCommands: map[string]string{"PreToolUse": "touch '" + pre + "'", "PostToolUse": "touch '" + post + "'"}}, harness.Hooks{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "artifact.txt"))
	if err != nil || string(data) != "verified\n" {
		t.Fatalf("workspace output %q: %v", data, err)
	}
	for _, f := range []string{pre, post} {
		if _, err = os.Stat(f); err != nil {
			t.Fatal(err)
		}
	}
}
