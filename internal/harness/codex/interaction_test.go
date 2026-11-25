package codex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func TestApprovalMapping(t *testing.T) {
	raw := json.RawMessage(`{"command":"example --one","cwd":"/work","availableDecisions":["accept","decline"]}`)
	for _, tt := range []struct {
		action string
		yes    bool
		want   string
	}{{"accept", true, "accept"}, {"accept", false, "decline"}, {"decline", true, "decline"}} {
		result, _, err := interaction(t.Context(), "item/commandExecution/requestApproval", raw, func(_ context.Context, msg string, schema map[string]any) (harness.Answer, error) {
			details := schema["properties"].(map[string]any)["approve"].(map[string]any)["description"].(string)
			if !strings.Contains(details, "example --one") || !strings.Contains(details, "/work") {
				t.Fatal(msg)
			}
			return harness.Answer{Action: tt.action, Content: map[string]any{"approve": tt.yes}}, nil
		})
		if err != nil || result.(map[string]string)["decision"] != tt.want {
			t.Fatalf("%v %v", result, err)
		}
	}
}

func TestLongFileApprovalDisplaysEntirePatch(t *testing.T) {
	diff := strings.Repeat("+line\n", 500) + "+last line"
	raw, err := json.Marshal(map[string]any{"itemId": "patch", "changes": []map[string]string{{"path": "example.txt", "diff": diff}}})
	if err != nil {
		t.Fatal(err)
	}
	result, _, err := interaction(t.Context(), "item/fileChange/requestApproval", raw, func(_ context.Context, heading string, schema map[string]any) (harness.Answer, error) {
		if len(heading) > 1024 {
			t.Fatal("oversized heading")
		}
		details := schema["properties"].(map[string]any)["approve"].(map[string]any)["description"].(string)
		var request struct{ Changes []struct{ Path, Diff string } }
		if err := json.Unmarshal([]byte(details), &request); err != nil || len(request.Changes) != 1 || request.Changes[0].Path != "example.txt" || request.Changes[0].Diff != diff {
			t.Fatal("incomplete approval details")
		}
		return harness.Answer{Action: "accept", Content: map[string]any{"approve": true}}, nil
	})
	if err != nil || result.(map[string]string)["decision"] != "accept" {
		t.Fatal(result, err)
	}
}

func TestAdditionalPermissionScopeMatchesApprovalForm(t *testing.T) {
	raw := json.RawMessage(`{"permissions":{"fileSystem":{"write":["/work/project"]}}}`)
	response, _, err := interaction(t.Context(), "item/permissions/requestApproval", raw, func(_ context.Context, _ string, schema map[string]any) (harness.Answer, error) {
		field := schema["properties"].(map[string]any)["approve"].(map[string]any)
		if !strings.Contains(field["title"].(string), "rest of this turn") || !strings.Contains(field["description"].(string), "/work/project") {
			t.Fatal("missing permission scope", field)
		}
		return harness.Answer{Action: "accept", Content: map[string]any{"approve": true}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got := response.(map[string]any)
	if got["scope"] != "turn" || string(got["permissions"].(json.RawMessage)) != `{"fileSystem":{"write":["/work/project"]}}` {
		t.Fatalf("wrong grant %v", got)
	}
}
