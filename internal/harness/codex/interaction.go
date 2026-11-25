package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

// Translate one native server request; the original JSON-RPC id is kept by rpc.
func interaction(ctx context.Context, method string, raw json.RawMessage, ask harness.Ask) (any, func(bool), error) {
	var p struct {
		GrantRoot   string
		Changes     []json.RawMessage
		Permissions json.RawMessage
		Questions   []struct {
			ID, Question, Header string
			IsOther, IsSecret    bool
			Options              []struct{ Label, Description string }
		}
		AvailableDecisions []any
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, nil, err
	}
	props := map[string]any{}
	required := []string{}
	var msg strings.Builder
	switch method {
	case "item/tool/requestUserInput":
		for _, q := range p.Questions {
			if q.IsSecret {
				return nil, nil, errors.New("secret answers cannot be recorded in conversation history")
			}
			if q.ID == "" || props[q.ID] != nil {
				return nil, nil, errors.New("invalid question ID")
			}
			v := map[string]any{"type": "string", "title": q.Header}
			var options []string
			var description strings.Builder
			description.WriteString(q.Question)
			for _, o := range q.Options {
				options = append(options, o.Label)
				fmt.Fprintf(&description, "\n%s: %s", o.Label, o.Description)
			}
			if !q.IsOther && len(options) > 0 {
				v["enum"] = options
			}
			v["description"] = description.String()
			props[q.ID] = v
			required = append(required, q.ID)
			fmt.Fprintln(&msg, q.Question)
		}
	case "item/commandExecution/requestApproval", "item/fileChange/requestApproval", "item/permissions/requestApproval":
		if method == "item/fileChange/requestApproval" && (p.GrantRoot != "" || len(p.Changes) == 0) {
			return nil, nil, errors.New("one-time file approval requires the proposed changes without a session grant")
		}
		// AutoWonder truncates the card heading to 1024 characters. The
		// question field retains its full description and the UI renders it.
		title := "Approve this request once?"
		if method == "item/permissions/requestApproval" {
			title = "Grant these permissions for the rest of this turn?"
		}
		fmt.Fprintf(&msg, "Codex %s", method)
		var details bytes.Buffer
		if err := json.Indent(&details, raw, "", "  "); err != nil {
			return nil, nil, err
		}
		props["approve"] = map[string]any{"type": "boolean", "title": title, "description": details.String()}
		required = []string{"approve"}
	default:
		return nil, nil, errors.New("unsupported Codex interaction")
	}
	a, err := ask(ctx, msg.String(), map[string]any{"type": "object", "properties": props, "required": required})
	if err != nil {
		return nil, nil, err
	}
	accepted := a.Action == "accept"
	if method == "item/tool/requestUserInput" {
		answers := map[string]any{}
		if accepted {
			for _, k := range required {
				v, ok := a.Content[k].(string)
				if !ok || v == "" {
					return nil, a.Written, errors.New("missing answer")
				}
				answers[k] = map[string]any{"answers": []string{v}}
			}
		}
		return map[string]any{"answers": answers}, a.Written, nil
	}
	yes, _ := a.Content["approve"].(bool)
	approved := accepted && yes
	if method == "item/permissions/requestApproval" {
		permissions := json.RawMessage("{}")
		if approved && len(p.Permissions) > 0 {
			permissions = p.Permissions
		}
		return map[string]any{"permissions": permissions, "scope": "turn"}, a.Written, nil
	}
	decision := "decline"
	if approved {
		decision = "accept"
	}
	if a.Action == "cancel" {
		decision = "cancel"
	}
	if len(p.AvailableDecisions) > 0 {
		found := false
		for _, d := range p.AvailableDecisions {
			if v, ok := d.(string); ok && decision == v {
				found = true
			}
		}
		if !found {
			return nil, a.Written, errors.New("decision not offered by Codex")
		}
	}
	return map[string]string{"decision": decision}, a.Written, nil
}
