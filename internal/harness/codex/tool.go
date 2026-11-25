package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func platformDefinition() any {
	return map[string]any{"type": "function", "name": "autowonder", "description": "List the AutoWonder tools available in this turn, or invoke a listed tool by name. When a conversation offers autowonder.propose_platform_actions, use it for writes that require owner confirmation. Report a pending proposal to the user; AutoWonder executes the frozen actions after confirmation in its UI.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"operation": map[string]any{"type": "string", "enum": []string{"list", "call"}}, "name": map[string]string{"type": "string"}, "arguments": map[string]any{"type": "object", "additionalProperties": true}}, "required": []string{"operation"}, "additionalProperties": false}}
}
func platformCall(ctx context.Context, invoke harness.PlatformTool, p json.RawMessage) (string, error) {
	var q struct {
		Tool      string
		Arguments struct {
			Operation, Name string
			Arguments       map[string]any
		}
	}
	dec := json.NewDecoder(bytes.NewReader(p))
	dec.UseNumber() // Preserve platform IDs in frozen action arguments.
	if err := dec.Decode(&q); err != nil {
		return "", err
	}
	if q.Tool != "autowonder" || invoke == nil {
		return "", errors.New("unknown task tool")
	}
	switch q.Arguments.Operation {
	case "list":
		return invoke(ctx, "", nil)
	case "call":
		if q.Arguments.Name != "" {
			if q.Arguments.Arguments == nil {
				q.Arguments.Arguments = map[string]any{}
			}
			return invoke(ctx, q.Arguments.Name, q.Arguments.Arguments)
		}
	}
	return "", errors.New("expected list or a named tool call")
}
