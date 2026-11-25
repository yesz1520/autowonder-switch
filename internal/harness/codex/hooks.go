package codex

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

// Trust only the generated handlers for this invocation, using Codex's own
// hashes. Leave project and user hook trust decisions untouched.
func (b Backend) startTurnServer(ctx context.Context, q harness.Request) (*rpc, error) {
	if len(q.ToolHookCommands) == 0 {
		return start(ctx, b.Binary, q.Directory, q)
	}
	var args []string
	for _, event := range []string{"PreToolUse", "PostToolUse"} {
		if cmd := q.ToolHookCommands[event]; cmd != "" {
			args = append(args, "-c", "hooks."+event+"=[{hooks=[{type=\"command\",timeout=86400,command="+strconv.Quote(cmd)+"}]}]")
		}
	}
	probe, err := startConfigured(ctx, b.Binary, q.Directory, q, args)
	if err != nil {
		return nil, err
	}
	states, err := hookStates(ctx, probe, q, false)
	probe.close()
	if err != nil {
		return nil, err
	}
	args = append(args, "-c", "hooks.state={"+strings.Join(states, ",")+"}")
	r, err := startConfigured(ctx, b.Binary, q.Directory, q, args)
	if err != nil {
		return nil, err
	}
	if _, err = hookStates(ctx, r, q, true); err != nil {
		r.close()
		return nil, err
	}
	return r, nil
}

func hookStates(ctx context.Context, r *rpc, q harness.Request, trusted bool) ([]string, error) {
	var list struct {
		Data []struct {
			Cwd   string
			Hooks []struct {
				Key, Command, EventName, CurrentHash, TrustStatus, Source string
				Enabled                                                   bool
			}
		}
	}
	if err := r.call(ctx, "hooks/list", map[string]any{"cwds": []string{q.Directory}}, &list); err != nil {
		return nil, err
	}
	var states []string
	for event, cmd := range q.ToolHookCommands {
		found := false
		for _, entry := range list.Data {
			for _, h := range entry.Hooks {
				if entry.Cwd == q.Directory && h.Command == cmd && strings.EqualFold(event, h.EventName) && h.Source == "sessionFlags" && h.Enabled && h.CurrentHash != "" && (!trusted || h.TrustStatus == "trusted") {
					states = append(states, strconv.Quote(h.Key)+"={trusted_hash="+strconv.Quote(h.CurrentHash)+"}")
					found = true
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("Codex did not activate %s package hook", event)
		}
	}
	return states, nil
}

// SupportsToolHooks checks the installed app-server without starting a model turn.
func (b Backend) SupportsToolHooks(ctx context.Context, dir string) bool {
	if _, err := exec.LookPath("curl"); err != nil {
		return false
	}
	r, err := b.startTurnServer(ctx, harness.Request{Directory: dir, ToolHookCommands: map[string]string{"PreToolUse": "true # autowonder preflight before", "PostToolUse": "true # autowonder preflight after"}})
	if err != nil {
		return false
	}
	r.close()
	return true
}
