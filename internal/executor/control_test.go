package executor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type waiting struct {
	stub
	started chan struct{}
}

func (b *waiting) Run(ctx context.Context, spec harness.Request, h harness.Hooks) (harness.Reply, error) {
	if _, err := os.Stat(filepath.Join(spec.Directory, "project")); err == nil {
		if strings.Contains(spec.Prompt, "--- Previous dispatch ---") {
			b, err := os.ReadFile(filepath.Join(spec.Directory, "project", "in-flight"))
			if err != nil || string(b) != "not committed" || !strings.Contains(spec.Prompt, "partial reply") {
				return harness.Reply{}, fmt.Errorf("missing restored work")
			}
		}
		if err = os.WriteFile(filepath.Join(spec.Directory, "project", "in-flight"), []byte("not committed"), 0600); err != nil {
			return harness.Reply{}, err
		}
	}
	if h.Session != nil {
		if err := h.Session("native"); err != nil {
			return harness.Reply{}, err
		}
	}
	if h.Ready != nil {
		h.Ready(func(context.Context, string) error { return nil })
	}
	if h.Message != nil {
		h.Message("partial reply")
	}
	b.started <- struct{}{}
	<-ctx.Done()
	return harness.Reply{}, ctx.Err()
}
func TestPauseAndCancel(t *testing.T) {
	ctx, stop := context.WithTimeout(t.Context(), 8*time.Second)
	defer stop()
	fixture := newGitFixture(t)
	repoSpec, _ := json.Marshal(map[string]any{"repos": []repository{fixture.spec}})
	pkg := archive(t, map[string]string{"workitem.md": "wait", "repos.json": string(repoSpec)})
	resumePkg := dispatchArchive(t, "2", map[string]string{"workitem.md": "continue", "repos.json": string(repoSpec)})
	recovery := make(chan []byte, 1)
	socket := make(chan *websocket.Conn, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/executor", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		socket <- c
		<-ctx.Done()
		c.CloseNow()
	})
	mux.HandleFunc("/package", func(w http.ResponseWriter, r *http.Request) { w.Write(pkg) })
	mux.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) { w.Write(resumePkg) })
	var posts atomic.Int32
	mux.HandleFunc("/api/daemon/dispatches/1/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		if posts.Add(1) == 1 {
			w.WriteHeader(503)
			return
		}
		f, _, err := r.FormFile("checkpoint")
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()
		data, _ := io.ReadAll(f)
		recovery <- data
		if r.FormValue("purpose") != "RECOVERY" || r.FormValue("providerSessionId") != "native" {
			t.Error("wrong recovery metadata")
		}
		json.NewEncoder(w).Encode(aw.CheckpointReceipt{CheckpointSeq: 1, SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(data)), SizeBytes: int64(len(data)), Purpose: "RECOVERY"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	dir := t.TempDir()
	backend := &waiting{started: make(chan struct{}, 1)}
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, aw.Config{BaseURL: srv.URL, ExecutorID: 7, Token: "test", Capacity: 2}, dir, backend, nil)
	}()
	defer func() { stop(); <-done }()
	var c *websocket.Conn
	select {
	case c = <-socket:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	send := func(m any) {
		t.Helper()
		b, _ := json.Marshal(m)
		if err := c.Write(ctx, websocket.MessageText, b); err != nil {
			t.Fatal(err)
		}
	}
	read := func(want string) map[string]any {
		t.Helper()
		for {
			_, b, err := c.Read(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			json.Unmarshal(b, &m)
			if m["type"] == want {
				return m
			}
		}
	}
	ready := func() {
		t.Helper()
		select {
		case <-backend.started:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	read("HEARTBEAT")
	send(map[string]any{"type": "TASK_DISPATCH", "executorId": 7, "dispatchId": 1, "attempt": 1, "packageId": "p", "downloadUrl": srv.URL + "/package", "checksumAlgorithm": "sha256", "checksumScope": "zip_archive", "checksum": fmt.Sprintf("sha256:%x", sha256.Sum256(pkg))})
	read("TASK_ACK")
	ready()
	send(map[string]any{"type": "TASK_GUIDANCE", "dispatchId": 1, "guidanceId": 77, "content": "change", "mode": "SAFE_STEER"})
	if read("TASK_GUIDANCE_ACK")["status"] != "APPLIED" {
		t.Fatal("guidance failed")
	}
	send(map[string]any{"type": "TASK_PAUSE", "dispatchId": 1})
	paused := read("TASK_PAUSED")
	if paused["checkpointSeq"] != float64(1) {
		t.Fatal(paused)
	}
	path := filepath.Join(dir, "restore-test")
	if err := os.WriteFile(path, <-recovery, 0600); err != nil {
		t.Fatal(err)
	}
	text, err := restoreAnswer(path, 1)
	if err != nil || text != "partial reply" {
		t.Fatalf("restore: %q %v", text, err)
	}
	_, fresh := prepareFixture(t, fixture.pkg)
	if _, err = restoreRepos(ctx, path, 1, fresh); err != nil {
		t.Fatal(err)
	}
	partial, err := os.ReadFile(filepath.Join(fresh[0].workDir, "in-flight"))
	if err != nil || string(partial) != "not committed" {
		t.Fatalf("paused edit lost: %q %v", partial, err)
	}
	send(map[string]any{"type": "TASK_RESULT_ACK", "dispatchId": 1, "accepted": true, "pauseCheckpointConfirmed": true, "checkpointSeq": 1, "checkpointSha256": paused["checkpointSha256"]})
	send(map[string]any{"type": "TASK_PAUSE", "dispatchId": 1})
	if again := read("TASK_PAUSED"); again["checkpointSha256"] != paused["checkpointSha256"] || posts.Load() != 2 {
		t.Fatalf("duplicate pause: %v uploads %d", again, posts.Load())
	}
	send(map[string]any{"type": "TASK_DISPATCH", "executorId": 7, "dispatchId": 2, "attempt": 1, "packageId": "p", "downloadUrl": srv.URL + "/resume", "checksumAlgorithm": "sha256", "checksumScope": "zip_archive", "checksum": fmt.Sprintf("sha256:%x", sha256.Sum256(resumePkg)), "resumeFromDispatchId": 1, "resumeCheckpointUrl": srv.URL + "/expired", "resumeCheckpointSha256": paused["checkpointSha256"]})
	read("TASK_ACK")
	ready()
	send(map[string]any{"type": "TASK_PAUSE", "dispatchId": 2, "cancel": true})
	send(map[string]any{"type": "CONVERSATION_TURN", "conversationId": 3, "turnId": 4, "dispatchAttempt": 2, "conversationDeliveryProtocolVersion": 1, "content": "wait"})
	ready()
	send(map[string]any{"type": "CONVERSATION_TURN_CANCEL", "conversationId": 3, "turnId": 4, "dispatchAttempt": 1})
	send(map[string]any{"type": "CONVERSATION_TURN_CANCEL", "conversationId": 3, "turnId": 4, "dispatchAttempt": 2})
	if got := read("CONVERSATION_TURN_ACK"); got["status"] != "CANCELED" || got["dispatchAttempt"] != float64(2) {
		t.Fatal(got)
	}
}

func TestGuidanceReceiptsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	count := 0
	r := runtime{dir: dir, tasks: map[int64]*job{1: {steer: func(context.Context, string) error {
		count++
		return nil
	}}}}
	check := func(r *runtime, kind, content, expected string) {
		t.Helper()
		status, message := r.applyOnce(t.Context(), kind, 1, 77, content)
		if status != expected {
			t.Fatalf("got %s (%s), want %s", status, message, expected)
		}
	}
	check(&r, "TASK_STEER", "first", "APPLIED")
	check(&r, "TASK_GUIDANCE", "second", "APPLIED")
	restarted := runtime{dir: dir}
	check(&restarted, "TASK_STEER", "first", "APPLIED")
	check(&r, "TASK_STEER", "changed", "FAILED")
	if err := os.WriteFile(filepath.Join(dir, "guidance", "1", "TASK_STEER", "77"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	check(&r, "TASK_STEER", "first", "FAILED")
	if count != 2 {
		t.Fatalf("injected %d times", count)
	}
}
