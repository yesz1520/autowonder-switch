package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

const handoffTool = "executor.request_handoff"

type handoffTarget struct {
	To     string `json:"to"`
	ToType string `json:"toType"`
}

type handoffProposal struct {
	WorkitemID int64
	Target     handoffTarget
}

// The proposal takes effect only when the platform accepts the successful result.
// Persist it before replying to Codex so delivery recovery can use the same target.
func withHandoff(platform harness.PlatformTool, dir string, workitemID int64) harness.PlatformTool {
	return func(ctx context.Context, name string, args map[string]any) (string, error) {
		switch name {
		case "":
			return appendTool(ctx, platform, json.RawMessage(`{"name":"executor.request_handoff","description":"Choose the next owner when the work item calls for a handoff. The platform routes the handoff after successful completion. Calling again replaces the proposal.","inputSchema":{"type":"object","properties":{"to":{"type":"string","description":"Target role or human from the work item roster"},"toType":{"type":"string","enum":["AGENT","HUMAN"]}},"required":["to","toType"],"additionalProperties":false}}`))
		case handoffTool:
			to, _ := args["to"].(string)
			kind, _ := args["toType"].(string)
			to = strings.TrimSpace(to)
			if to == "" || (kind != "HUMAN" && kind != "AGENT") {
				return "", errors.New("handoff requires a target and toType AGENT or HUMAN")
			}
			if err := save(filepath.Join(dir, "handoff-proposal.json"), handoffProposal{workitemID, handoffTarget{to, kind}}); err != nil {
				return "", err
			}
			return `{"status":"PROPOSED"}`, nil
		default:
			if platform == nil {
				return "", errors.New("unknown platform tool")
			}
			return platform(ctx, name, args)
		}
	}
}

func (r *runtime) handoffResult(_ context.Context, _ *aw.Session, m *aw.TaskHandoffResult) error {
	if m.DispatchID <= 0 || m.WorkitemID <= 0 {
		return nil
	}
	dir := filepath.Join(r.dir, fmt.Sprintf("task-%d", m.DispatchID))
	var p handoffProposal
	ok, err := cached(filepath.Join(dir, "handoff-proposal.json"), &p)
	if err != nil || !ok || p.WorkitemID != m.WorkitemID {
		return err
	}
	return save(filepath.Join(dir, "handoff-result.json"), m)
}
