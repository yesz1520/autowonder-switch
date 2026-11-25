package codex

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFrozenActionArguments(t *testing.T) {
	raw := json.RawMessage(`{"tool":"autowonder","arguments":{"operation":"call","name":"autowonder.propose_platform_actions","arguments":{"actions":[{"tool":"autowonder.update_workitem","arguments":{"id":9007199254740993}}]}}}`)
	reply, err := platformCall(t.Context(), func(_ context.Context, name string, args map[string]any) (string, error) {
		if name != "autowonder.propose_platform_actions" {
			t.Fatal(name)
		}
		b, err := json.Marshal(args)
		if err != nil || string(b) != `{"actions":[{"arguments":{"id":9007199254740993},"tool":"autowonder.update_workitem"}]}` {
			t.Fatalf("%s %v", b, err)
		}
		return `{"structuredContent":{"id":"22","status":"PENDING"}}`, nil
	}, raw)
	if err != nil || reply == "" {
		t.Fatal(reply, err)
	}
}
