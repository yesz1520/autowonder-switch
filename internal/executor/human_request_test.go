package executor

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type reviewTurn struct{ stub }

func (b *reviewTurn) Run(ctx context.Context, q harness.Request, _ harness.Hooks) (harness.Reply, error) {
	n := b.calls.Add(1)
	if n == 1 {
		if q.Ask == nil {
			return harness.Reply{}, fmt.Errorf("missing input callback")
		}
		ans, err := q.Ask(ctx, "Which branch?", map[string]any{"type": "object", "properties": map[string]any{"branch": map[string]string{"type": "string"}}})
		if err != nil || ans.Action != "accept" || ans.Content["branch"] != "main" {
			return harness.Reply{}, fmt.Errorf("answer=%v err=%v", ans, err)
		}
	} else if q.Session != "session" || (n == 2 && !strings.Contains(q.Prompt, "fix typo")) || (n == 3 && q.Ask != nil) {
		return harness.Reply{}, fmt.Errorf("unexpected next step %d", n)
	}
	return harness.Reply{Session: "session", Text: fmt.Sprintf("draft %d", n), Usage: &harness.Usage{Provider: "test", Model: "sample", Input: 7, Output: 3}}, nil
}
func TestWorkflowHumanRequests(t *testing.T) {
	data := archive(t, map[string]string{"sdlc.json": `{"currentStepId":"1","steps":[{"id":"1","name":"Review","retryBudget":1,"gatePolicy":{"askWhenUnclear":true,"humanApprovalRequired":true}},{"id":"2","name":"Finish"}]}`})
	file := filepath.Join(t.TempDir(), "package.zip")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	var usageCalls atomic.Int32
	mux.HandleFunc("POST /api/daemon/tasks/7/usage", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Usage []aw.UsageEntry }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Usage) != 1 {
			t.Error("invalid usage")
			return
		}
		n := usageCalls.Add(1)
		step, turns := "1", int64(n)
		if n == 3 {
			step, turns = "2", 1
		}
		entry := body.Usage[0]
		if entry.StepID != step || entry.InputTokens == nil || *entry.InputTokens != 7*turns || entry.OutputTokens == nil || *entry.OutputTokens != 3*turns {
			t.Errorf("step %s expected %d turns, got %+v", step, turns, entry)
		}
		fmt.Fprint(w, "{}")
	})
	var pending aw.HumanRequest
	mux.HandleFunc("POST /api/daemon/dispatches/7/requests", func(w http.ResponseWriter, r *http.Request) {
		var in aw.HumanRequestInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Error(err)
		}
		if r.Header.Get("Authorization") != "Bearer test" || in.StepID != "1" || in.Key == "" {
			t.Errorf("invalid request %+v", in)
		}
		out := aw.HumanRequest{DispatchID: 7, Key: in.Key, Kind: in.Kind, ID: 9, Status: "RESOLVED"}
		if in.Kind == "INPUT" {
			out.Status = "PENDING"
			pending = out
		} else {
			var body struct{ Fingerprint string }
			json.Unmarshal(in.Context, &body)
			hash, err := hex.DecodeString(body.Fingerprint)
			if err != nil || len(hash) != 32 {
				t.Error("invalid fingerprint")
			}
			if in.StepAttempt == 1 {
				out.Answer = json.RawMessage(`{"action":"reject","content":{"note":"fix typo"}}`)
			} else if in.StepAttempt == 2 {
				out.Answer = json.RawMessage(`{"action":"approve"}`)
			} else {
				t.Error("unexpected step attempt")
			}
		}
		json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("GET /api/daemon/dispatches/7/requests/9", func(w http.ResponseWriter, r *http.Request) {
		out := pending
		out.Status = "RESOLVED"
		out.Answer = json.RawMessage(`{"action":"accept","content":{"branch":"main"}}`)
		json.NewEncoder(w).Encode(out)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := aw.New(aw.Config{BaseURL: server.URL, ExecutorID: 1, Token: "test"})
	if err != nil {
		t.Fatal(err)
	}
	backend := &reviewTurn{}
	r := runtime{client: client, backend: backend}
	out, err := r.runWorkflow(t.Context(), &aw.TaskDispatch{DispatchID: 7, Attempt: 1}, file, harness.Request{}, harness.Hooks{}, nil, nil)
	if err != nil || out.Text != "draft 3" || backend.calls.Load() != 3 || usageCalls.Load() != 3 {
		t.Fatalf("result=%+v err=%v", out, err)
	}
}

func TestInstalledCodexDispatchQuestion(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires Codex login")
	}
	ctx, stop := context.WithTimeout(t.Context(), 90*time.Second)
	defer stop()
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.zip")
	data := archive(t, map[string]string{"sdlc.json": `{"currentStepId":"7","steps":[{"id":"7","name":"Choose","gatePolicy":{"askWhenUnclear":true}}]}`})
	if err := os.WriteFile(pkg, data, 0600); err != nil {
		t.Fatal(err)
	}
	var asked, metered atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/daemon/tasks/7/usage" {
			var body struct{ Usage []aw.UsageEntry }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Usage) != 1 {
				t.Error("missing usage entry")
				return
			}
			entry := body.Usage[0]
			if entry.StepID != "7" || entry.Model == "" || entry.Provider == "" || entry.OutputTokens == nil || entry.ReasoningTokens == nil || *entry.OutputTokens+*entry.ReasoningTokens <= 0 {
				t.Errorf("invalid native usage %+v", entry)
			}
			metered.Store(true)
			fmt.Fprint(w, "{}")
			return
		}
		var in aw.HumanRequestInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Error(err)
			return
		}
		var schema struct{ Required []string }
		if err := json.Unmarshal(in.Schema, &schema); err != nil || len(schema.Required) == 0 {
			t.Errorf("invalid form: %s", in.Schema)
			return
		}
		content := map[string]string{}
		for _, key := range schema.Required {
			content[key] = "teal"
		}
		answer, _ := json.Marshal(map[string]any{"action": "accept", "content": content})
		asked.Store(true)
		json.NewEncoder(w).Encode(aw.HumanRequest{ID: 9, DispatchID: 7, Kind: in.Kind, Key: in.Key, Status: "RESOLVED", Answer: answer})
	}))
	defer server.Close()
	c, err := aw.New(aw.Config{BaseURL: server.URL, ExecutorID: 1, Token: "test"})
	if err != nil {
		t.Fatal(err)
	}
	r := runtime{client: c, backend: codex.Backend{Binary: "codex"}}
	reply, err := r.runWorkflow(ctx, &aw.TaskDispatch{DispatchID: 7, Attempt: 1}, pkg, harness.Request{Directory: dir, Prompt: "Use request_user_input to ask me which color to use, teal or purple. When the user answers, reply with only their chosen color."}, harness.Hooks{}, nil, nil)
	if err != nil || !asked.Load() || !metered.Load() || !strings.Contains(strings.ToLower(reply.Text), "teal") {
		t.Fatalf("asked=%v answer=%q err=%v", asked.Load(), reply.Text, err)
	}
}

func TestCancelPendingHumanRequest(t *testing.T) {
	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	var calls atomic.Int32
	var canceled atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/daemon/dispatches/4/requests", func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(503)
			return
		}
		var in aw.HumanRequestInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Key != "same" {
			t.Errorf("payload %v err=%v", in, err)
		}
		json.NewEncoder(w).Encode(aw.HumanRequest{ID: 3, DispatchID: 4, Key: "same", Kind: "INPUT", Status: "PENDING"})
	})
	mux.HandleFunc("GET /api/daemon/dispatches/4/requests/3", func(w http.ResponseWriter, r *http.Request) { stop() })
	mux.HandleFunc("POST /api/daemon/dispatches/4/requests/3/cancel", func(w http.ResponseWriter, r *http.Request) {
		canceled.Store(true)
		fmt.Fprint(w, `{"status":"CANCELED"}`)
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	c, err := aw.New(aw.Config{ExecutorID: 1, Token: "test", BaseURL: s.URL})
	if err != nil {
		t.Fatal(err)
	}
	r := runtime{client: c}
	_, err = r.humanRequest(ctx, 4, aw.HumanRequestInput{Kind: "INPUT", Key: "same"})
	if !errors.Is(err, context.Canceled) || !canceled.Load() || calls.Load() != 2 {
		t.Fatalf("cancel=%v calls=%v err=%v", canceled.Load(), calls.Load(), err)
	}
}
