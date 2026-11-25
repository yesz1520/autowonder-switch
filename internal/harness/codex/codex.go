// Package codex uses the installed Codex app-server and local login.
package codex

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type Backend struct{ Binary string }

func (b Backend) Executable() string { return b.Binary }

func (b Backend) Version(ctx context.Context) (string, error) {
	p, err := exec.CommandContext(ctx, b.Binary, "--version").Output()
	if err != nil {
		return "", err
	}
	v := strings.Fields(string(p))
	if len(v) != 2 || v[0] != "codex-cli" {
		return "", errors.New("unexpected Codex version")
	}
	return v[1], nil
}
func (b Backend) Models(ctx context.Context) (harness.Catalog, error) {
	var cat harness.Catalog
	r, err := start(ctx, b.Binary, "", harness.Request{})
	if err != nil {
		return cat, err
	}
	defer r.close()
	var cfg struct{ Config struct{ Model string } }
	if err = r.call(ctx, "config/read", map[string]any{}, &cfg); err != nil {
		return cat, err
	}
	cat.Default = cfg.Config.Model
	params := map[string]any{}
	seen := map[string]bool{}
	for {
		var page struct {
			Data []struct {
				ID, Model, DisplayName string
				IsDefault              bool
			}
			NextCursor string
		}
		if err = r.call(ctx, "model/list", params, &page); err != nil {
			return cat, err
		}
		for _, v := range page.Data {
			id := v.Model
			if id == "" {
				id = v.ID
			}
			if id == "" {
				continue
			}
			if v.IsDefault && cat.Default == "" {
				cat.Default = id
			}
			name := v.DisplayName
			if name == "" {
				name = id
			}
			cat.Models = append(cat.Models, harness.Model{ID: id, Name: name})
		}
		if page.NextCursor == "" {
			return cat, nil
		}
		if seen[page.NextCursor] {
			return cat, errors.New("repeated Codex catalog cursor")
		}
		seen[page.NextCursor] = true
		params["cursor"] = page.NextCursor
	}
}
func (b Backend) Run(ctx context.Context, q harness.Request, h harness.Hooks) (reply harness.Reply, err error) {
	r, err := b.startTurnServer(ctx, q)
	if err != nil {
		return reply, err
	}
	defer r.close()
	var cfg struct {
		Config struct {
			Plugins map[string]json.RawMessage
			Servers map[string]json.RawMessage `json:"mcp_servers"`
		}
	}
	if err = r.call(ctx, "config/read", map[string]string{"cwd": q.Directory}, &cfg); err != nil {
		return reply, err
	}
	servers := map[string]any{}
	for name := range cfg.Config.Servers {
		servers[name] = map[string]bool{"enabled": false}
	}
	overrides := map[string]any{
		"mcp_servers": servers, "web_search": "disabled",
		"features.shell_tool": q.Readable || q.Writable, "features.unified_exec": q.Readable || q.Writable,
		"features.plugins": len(q.PluginIDs) > 0, "features.browser_use": false, "features.in_app_browser": false,
		"sandbox_workspace_write.network_access": false,
	}
	if q.MemoryMode != "provider-local" {
		overrides["memories.use_memories"] = false
		overrides["memories.generate_memories"] = false
	}
	if len(q.PluginIDs) > 0 {
		enabled := map[string]any{"_default": map[string]bool{"enabled": false}}
		for id := range cfg.Config.Plugins {
			if id != "_default" {
				enabled[id] = map[string]bool{"enabled": false}
			}
		}
		for _, id := range q.PluginIDs {
			enabled[id] = map[string]bool{"enabled": true}
		}
		overrides["plugins"] = enabled
		overrides["features.remote_plugin"] = false
	}
	p := map[string]any{"cwd": q.Directory, "sandbox": "read-only", "approvalPolicy": "never", "config": overrides}
	if q.Ask != nil {
		overrides["features.default_mode_request_user_input"] = true
	}
	if q.Approvals && q.Ask != nil {
		p["approvalPolicy"] = "on-request"
		p["approvalsReviewer"] = "user"
		overrides["features.request_permissions_tool"] = true
	}
	if q.Writable {
		p["sandbox"] = "workspace-write"
	}
	// Dynamic tool definitions belong to the thread. Later turns can have different credentials.
	if q.Session == "" {
		p["dynamicTools"] = []any{platformDefinition()}
	}
	if len(q.Environment) > 0 {
		overrides["shell_environment_policy.set"] = q.Environment
	}
	method := "thread/start"
	if q.Session != "" {
		method = "thread/resume"
		if q.Fork {
			method = "thread/fork"
		}
		p["threadId"] = q.Session
	}
	if q.Model != "" {
		p["model"] = q.Model
	}
	if q.Instructions != "" {
		p["developerInstructions"] = q.Instructions
	}
	var thread struct {
		Thread               struct{ ID string }
		Model, ModelProvider string
	}
	if err = r.call(ctx, method, p, &thread); err != nil {
		return reply, err
	}
	if thread.Thread.ID == "" || q.Session != "" && ((q.Fork && q.Session == thread.Thread.ID) || (!q.Fork && q.Session != thread.Thread.ID)) || q.Model != "" && q.Model != thread.Model {
		return reply, errors.New("Codex returned unexpected session or model")
	}
	reply.Session = thread.Thread.ID
	reply.Model = thread.Model
	if h.Session != nil {
		if err = h.Session(reply.Session); err != nil {
			return reply, err
		}
	}
	input := func(text string) []map[string]string { return []map[string]string{{"type": "text", "text": text}} }
	var turn struct{ Turn struct{ ID string } }
	items := input(q.Prompt)
	for _, path := range q.Images {
		items = append(items, map[string]string{"type": "localImage", "path": path})
	}
	params := map[string]any{"threadId": reply.Session, "input": items}
	if q.OutputSchema != nil {
		params["outputSchema"] = q.OutputSchema
	}
	if err = r.call(ctx, "turn/start", params, &turn); err != nil {
		return reply, err
	}
	if turn.Turn.ID == "" {
		return reply, errors.New("missing Codex turn ID")
	}
	if h.TurnStarted != nil {
		if err = h.TurnStarted(reply.Session, turn.Turn.ID); err != nil {
			return reply, err
		}
	}
	if h.Ready != nil {
		tid := reply.Session
		h.Ready(func(c context.Context, text string) error {
			return r.call(c, "turn/steer", map[string]any{"threadId": tid, "expectedTurnId": turn.Turn.ID, "input": input(text)}, nil)
		})
	}
	var lastTotal int64
	for {
		select {
		case <-ctx.Done():
			return reply, ctx.Err()
		case <-r.done:
			return reply, errors.New("Codex disconnected before completing turn")
		case f := <-r.events:
			var ev struct {
				ThreadID, TurnID string
				TokenUsage       tokenUpdate
				Item             struct{ Type, Text string }
				Turn             struct {
					ID, Status string
					Error      *struct{ Message string }
				}
			}
			if err = json.Unmarshal(f.Params, &ev); err != nil {
				return reply, err
			}
			if ev.ThreadID != reply.Session {
				continue
			}
			if f.Method == "thread/tokenUsage/updated" && ev.TurnID == turn.Turn.ID {
				if ev.TokenUsage.Total.TotalTokens > lastTotal {
					lastTotal = ev.TokenUsage.Total.TotalTokens
					if reply.Usage == nil {
						reply.Usage = &harness.Usage{Provider: thread.ModelProvider, Model: reply.Model}
					}
					ev.TokenUsage.addLast(reply.Usage)
				}
			}
			if f.Method == "item/completed" && ev.TurnID == turn.Turn.ID && ev.Item.Type == "agentMessage" {
				reply.Text = ev.Item.Text
				if h.Message != nil {
					h.Message(reply.Text)
				}
			}
			if f.Method == "turn/completed" && ev.Turn.ID == turn.Turn.ID {
				if ev.Turn.Status != "completed" {
					if ev.Turn.Error != nil {
						return reply, errors.New(ev.Turn.Error.Message)
					}
					return reply, errors.New("Codex turn " + ev.Turn.Status)
				}
				if strings.TrimSpace(reply.Text) == "" {
					return reply, errors.New("Codex produced no answer")
				}
				return reply, nil
			}
		}
	}
}
