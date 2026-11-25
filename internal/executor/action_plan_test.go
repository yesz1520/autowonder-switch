package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func TestInstalledCodexProposesFrozenPlan(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login")
	}
	ctx, stop := context.WithTimeout(t.Context(), 100*time.Second)
	defer stop()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mcp" || r.Header.Get("Authorization") != "Bearer turn-test" {
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			w.WriteHeader(403)
			return
		}
		var req struct {
			Method string
			Params struct {
				Name      string
				Arguments struct {
					Actions []struct {
						Tool      string
						Arguments struct {
							Title string
							ID    json.Number
						}
					}
				}
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if req.Method == "tools/list" {
			fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"autowonder.propose_platform_actions","description":"Submit actions for owner confirmation","inputSchema":{"type":"object","required":["actions"],"properties":{"actions":{"type":"array"}}}}]}}`)
			return
		}
		a := req.Params.Arguments.Actions
		if req.Method != "tools/call" || req.Params.Name != "autowonder.propose_platform_actions" || len(a) != 1 || a[0].Tool != "autowonder.update_workitem" || a[0].Arguments.ID.String() != "9007199254740993" || a[0].Arguments.Title != "Violet" {
			t.Errorf("unexpected action %+v", req)
			w.WriteHeader(400)
			return
		}
		calls.Add(1)
		fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"result":{"structuredContent":{"id":"plan-813","status":"PENDING","payloadHash":"hash813"},"isError":false}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	work := filepath.Join(dir, "workspace")
	if err := os.Mkdir(work, 0700); err != nil {
		t.Fatal(err)
	}
	r := runtime{base: server.URL, backend: codex.Backend{Binary: "codex"}}
	r.Endpoint, _ = aw.NewEndpoint(r.snapshot, nil)
	turn := aw.ConversationTurn{ConversationID: 3, TurnID: 8, DispatchAttempt: 1, MCPToken: "turn-test", Content: "Using the autowonder tool, propose exactly one action for owner confirmation: autowonder.update_workitem with arguments id (integer) 9007199254740993 and title Violet. Report the returned plan ID and status. Do not execute the action."}
	reply, err := r.converse(&job{ctx: ctx}, &turn, dir, work)
	if err != nil || calls.Load() != 1 || !strings.Contains(reply.Text, "plan-813") {
		t.Fatalf("calls=%d reply=%q err=%v", calls.Load(), reply.Text, err)
	}
}
