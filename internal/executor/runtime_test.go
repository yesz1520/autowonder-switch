package executor

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/claude"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
	"github.com/yesz1520/autowonder-switch/internal/harness/pi"
)

type stub struct{ calls atomic.Int32 }

func (*stub) Version(context.Context) (string, error) { return "test", nil }
func (*stub) Models(context.Context) (harness.Catalog, error) {
	return harness.Catalog{Default: "model", Models: []harness.Model{{ID: "model", Name: "Model"}}}, nil
}
func (b *stub) Run(_ context.Context, q harness.Request, h harness.Hooks) (harness.Reply, error) {
	b.calls.Add(1)
	if h.Session != nil {
		if err := h.Session("session"); err != nil {
			return harness.Reply{}, err
		}
	}
	if h.Message != nil {
		h.Message("violet")
	}
	return harness.Reply{Session: "session", Text: "violet", Model: "model"}, nil
}
func archive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	return dispatchArchive(t, "1", files)
}
func dispatchArchive(t *testing.T, dispatch string, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	m := aw.Manifest{SchemaVersion: "autoWonder.taskPackage.v1", PackageID: "p", DispatchID: dispatch, FileDigests: map[string]string{}}
	for name, data := range files {
		f, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.WriteString(f, data); err != nil {
			t.Fatal(err)
		}
		m.FileDigests[name] = fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(data)))
	}
	f, err := z.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(f).Encode(m); err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type handoffStub struct{ stub }

func (b *handoffStub) Run(ctx context.Context, q harness.Request, h harness.Hooks) (harness.Reply, error) {
	list, err := q.Platform(ctx, "", nil)
	if err != nil || !strings.Contains(list, handoffTool) {
		return harness.Reply{}, fmt.Errorf("handoff missing from catalog: %s (%v)", list, err)
	}
	if _, err = q.Platform(ctx, handoffTool, map[string]any{"to": "reviewer", "toType": "AGENT"}); err != nil {
		return harness.Reply{}, err
	}
	return b.stub.Run(ctx, q, h)
}
func TestDispatchRoundTrip(t *testing.T)        { runDispatchRoundTrip(t, false) }
func TestDeliverRecoveredTaskTurn(t *testing.T) { runDispatchRoundTrip(t, true) }
func runDispatchRoundTrip(t *testing.T, recovery bool) {
	b := &handoffStub{}
	var backend harness.Backend = b
	reader := &completedTurn{}
	savedID := aw.NewResultID()
	useClaude := !recovery && os.Getenv("AUTOWONDER_TEST_CLAUDE") == "1"
	usePi := !recovery && os.Getenv("AUTOWONDER_TEST_PI") == "1"
	provider := "codex"
	if recovery {
		backend = reader
	} else if usePi {
		piFixture(t)
		backend = pi.Backend{}
		provider = "pi"
	} else if useClaude {
		provider = "claude"
		claudeFixture(t)
		backend = claude.Backend{}
	} else if os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
		backend = codex.Backend{Binary: "codex"}
	}
	data := archive(t, map[string]string{"workitem.md": "Request a handoff to reviewer (AGENT) using the executor.request_handoff tool. Then reply with only the word violet."})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Second)
	defer cancel()
	connected := make(chan *websocket.Conn, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/executor", func(w http.ResponseWriter, r *http.Request) {
		c, e := websocket.Accept(w, r, nil)
		if e != nil {
			t.Error(e)
			return
		}
		connected <- c
		<-ctx.Done()
		c.CloseNow()
	})
	var usageCalls atomic.Int32
	mux.HandleFunc("/api/daemon/tasks/1/usage", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Usage []aw.UsageEntry }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Usage) == 0 {
			t.Error("missing usage")
			return
		}
		if (useClaude || usePi) && (req.Usage[0].Provider != provider || req.Usage[0].InputTokens == nil || *req.Usage[0].InputTokens < 1) {
			t.Error("incorrect harness usage")
		}
		usageCalls.Add(1)
		w.WriteHeader(204)
	})
	mux.HandleFunc("/package", func(w http.ResponseWriter, r *http.Request) { w.Write(data) })
	mux.HandleFunc("/api/daemon/dispatches/1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("files")
		if err != nil {
			t.Error(err)
			return
		}
		defer file.Close()
		v, _ := io.ReadAll(file)
		if !(recovery && string(v) == "recovered answer") && !strings.Contains(strings.ToLower(string(v)), "violet") {
			t.Errorf("artifact: %q", v)
		}
		fmt.Fprint(w, `{"files":[{"index":0,"status":"ACCEPTED"}]}`)
	})
	mux.HandleFunc("/api/daemon/dispatches/1/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		if (useClaude || usePi) && r.FormValue("provider") != provider {
			t.Error("incorrect checkpoint provider")
		}
		f, _, err := r.FormFile("checkpoint")
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()
		data, _ := io.ReadAll(f)
		json.NewEncoder(w).Encode(aw.CheckpointReceipt{CheckpointSeq: 1, SizeBytes: int64(len(data)), Purpose: "COMPLETION", ResultID: r.FormValue("resultId"), SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(data))})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	defer cancel()
	done := make(chan error, 1)
	dir := t.TempDir()
	if recovery {
		taskDir := filepath.Join(dir, fmt.Sprintf("%x", sha256.Sum256([]byte(server.URL+"/7"))), "task-1")
		if err := os.MkdirAll(filepath.Join(taskDir, "workspace"), 0700); err != nil {
			t.Fatal(err)
		}
		for name, bytes := range map[string][]byte{"package.zip": data, "started": {}, "native-turn.json": []byte(`{"sessionId":"thread-a","turnId":"turn-b"}`)} {
			if err := writeFile(filepath.Join(taskDir, name), bytes); err != nil {
				t.Fatal(err)
			}
		}
		state := taskRecovery{PackageID: "p", Checksum: fmt.Sprintf("sha256:%x", sha256.Sum256(data)), Attempt: 1, ResultID: savedID}
		if err := save(filepath.Join(taskDir, "task-recovery.json"), state); err != nil {
			t.Fatal(err)
		}
	}
	go func() {
		done <- Run(ctx, aw.Config{BaseURL: server.URL, ExecutorID: 7, Token: "example"}, dir, backend, nil)
	}()
	var c *websocket.Conn
	select {
	case c = <-connected:
	case e := <-done:
		t.Fatal(e)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	read := func(want string) []byte {
		t.Helper()
		for {
			_, v, e := c.Read(ctx)
			if e != nil {
				t.Fatal(e)
			}
			var m struct{ Type string }
			json.Unmarshal(v, &m)
			if m.Type == want {
				return v
			}
		}
	}
	send := func(v any) {
		t.Helper()
		m, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		if e = c.Write(ctx, websocket.MessageText, m); e != nil {
			t.Fatal(e)
		}
	}
	h := read("HEARTBEAT")
	if !bytes.Contains(h, []byte("aws-"+provider+"-")) {
		t.Fatal(string(h))
	}
	dispatch := map[string]any{"type": "TASK_DISPATCH", "executorId": 7, "dispatchId": 1, "workitemId": 23, "attempt": 1, "packageId": "p", "downloadUrl": server.URL + "/package", "checksumAlgorithm": "sha256", "checksumScope": "zip_archive", "checksum": fmt.Sprintf("sha256:%x", sha256.Sum256(data))}
	send(dispatch)
	read("TASK_ACK")
	var result aw.TaskResult
	json.Unmarshal(read("TASK_RESULT"), &result)
	if !result.Success || result.CheckpointSHA256 == "" {
		t.Fatalf("task: %+v", result)
	}
	if recovery && (result.ResultID != savedID || result.ResultSummary != "recovered answer" || reader.calls.Load() != 0) {
		t.Fatalf("recovered task %+v calls=%d", result, reader.calls.Load())
	}
	if !recovery {
		var target handoffTarget
		if err := json.Unmarshal(result.Handoff, &target); err != nil || target != (handoffTarget{"reviewer", "AGENT"}) {
			t.Fatalf("handoff %s %v", result.Handoff, err)
		}
		send(map[string]any{"type": "TASK_HANDOFF_RESULT", "workitemId": 23, "dispatchId": 1, "status": "HUMAN_ASSIGNED", "targetType": "HUMAN", "targetRef": 42})
	}
	send(dispatch)
	var duplicate aw.TaskResult
	json.Unmarshal(read("TASK_RESULT"), &duplicate)
	if !recovery {
		paths, err := filepath.Glob(filepath.Join(dir, "*", "task-1", "handoff-result.json"))
		if err != nil || len(paths) != 1 {
			t.Fatalf("handoff result files: %v %v", paths, err)
		}
		var decision aw.TaskHandoffResult
		if ok, e := cached(paths[0], &decision); !ok || e != nil || decision.TargetType != "HUMAN" || decision.TargetRef != 42 || !bytes.Equal(duplicate.Handoff, result.Handoff) || (backend == b && b.calls.Load() != 1) {
			t.Fatalf("replayed handoff: %+v %v", decision, e)
		}
	}
	if duplicate.ResultID != result.ResultID {
		t.Fatal("duplicate execution")
	}
	send(map[string]any{"type": "TASK_RESULT_ACK", "dispatchId": 1, "resultId": result.ResultID, "resultProtocolVersion": 2, "accepted": true, "disposition": "ACCEPTED"})
	if useClaude || usePi {
		if usageCalls.Load() != 1 {
			t.Fatal("expected one usage report")
		}
		var session string
		for i, prompt := range []string{"first-turn-marker", "confirm-prior-turn"} {
			send(map[string]any{"type": "CONVERSATION_TURN", "executorId": 7, "conversationId": 13, "turnId": i + 1, "dispatchAttempt": 1, "content": prompt, "cliSessionRef": session})
			var ack aw.ConversationTurnAck
			if err := json.Unmarshal(read("CONVERSATION_TURN_ACK"), &ack); err != nil || ack.Status != "SUCCESS" || ack.SessionID == "" || (i == 1 && (ack.SessionID != session || ack.ReplyMarkdown != "preserved-history")) {
				t.Fatalf("conversation %d: %+v %v", i, ack, err)
			}
			session = ack.SessionID
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("executor shutdown timed out")
	}
}
func TestCheckoutPatch(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	for _, a := range [][]string{{"init"}, {"-c", "user.name=t", "-c", "user.email=t@example.test", "commit", "--allow-empty", "-m", "initial"}} {
		cmd := exec.Command("git", append([]string{"-C", source}, a...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v", out, err)
		}
	}
	p := filepath.Join(root, "package.zip")
	os.WriteFile(p, archive(t, map[string]string{"workitem.md": "write", "repos.json": `{"repos":[{"name":"project"}]}`}), 0600)
	work := filepath.Join(root, "workspace")
	repos, err := prepareRepos(t.Context(), p, root, work, map[string]string{"project": source})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(work, "project", "new.txt"), []byte("new\n"), 0600); err != nil {
		t.Fatal(err)
	}
	patch, err := repos[0].patch(t.Context())
	if err != nil || !bytes.Contains(patch, []byte("+new\n")) {
		t.Fatalf("%s %v", patch, err)
	}
	if _, err = os.Stat(filepath.Join(source, "new.txt")); !os.IsNotExist(err) {
		t.Fatal("modified source")
	}
}
