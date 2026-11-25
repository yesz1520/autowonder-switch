package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestAcceptanceToolBindsCurrentDispatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" || req.URL.Path != "/api/daemon/dispatches/71/acceptance" || req.URL.Query().Get("token") != "local-key" {
			t.Error("incorrect dispatch or credential")
		}
		var reports []aw.AcceptanceReport
		if err := json.NewDecoder(req.Body).Decode(&reports); err != nil || len(reports) != 1 || reports[0].RequirementKey != "R1" {
			t.Errorf("unexpected reports %v: %v", reports, err)
		}
		fmt.Fprint(w, `{"dispatchId":71,"reported":1}`)
	}))
	defer server.Close()
	client, err := aw.New(aw.Config{BaseURL: server.URL, ExecutorID: 42, Token: "local-key"})
	if err != nil {
		t.Fatal(err)
	}
	r := runtime{client: client}
	tool := r.withAcceptance(withHandoff(nil, t.TempDir(), 1), 71)
	list, err := tool(context.Background(), "", nil)
	if err != nil || !strings.Contains(list, "executor.report_acceptance") || !strings.Contains(list, handoffTool) {
		t.Fatalf("tools %s: %v", list, err)
	}
	for _, args := range []map[string]any{nil, {"reports": "invalid"}} {
		if _, err := tool(context.Background(), "executor.report_acceptance", args); err == nil {
			t.Error("invalid report accepted")
		}
	}
	result, err := tool(context.Background(), "executor.report_acceptance", map[string]any{"dispatchId": 99, "reports": []aw.AcceptanceReport{{RequirementKey: "R1", Implementation: "PARTIAL", Verdict: "NOT_VERIFIED"}}})
	if err != nil || result != `{"dispatchId":71,"reported":1}` {
		t.Fatalf("receipt %s: %v", result, err)
	}
}
