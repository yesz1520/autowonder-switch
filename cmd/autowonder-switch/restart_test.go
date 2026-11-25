package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

// Exercise exec(2), argument/environment preservation, and the native Codex
// startup probes through the same binary that operators run.
func TestNativeProcessRestart(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" || (runtime.GOOS != "darwin" && runtime.GOOS != "linux") {
		t.Skip("requires the native Codex CLI and exec support")
	}
	ctx, stop := context.WithTimeout(t.Context(), 80*time.Second)
	defer stop()
	bin := filepath.Join(t.TempDir(), "autowonder-switch")
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %s %v", out, err)
	}
	sockets := make(chan *websocket.Conn, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws/executor" || r.URL.Query().Get("token") != "fixture" {
			http.Error(w, "invalid handshake", 401)
			return
		}
		c, e := websocket.Accept(w, r, nil)
		if e == nil {
			sockets <- c
		}
	}))
	defer srv.Close()
	cmd := exec.CommandContext(ctx, bin, "-executor-id", "7", "-url", srv.URL, "-state-dir", t.TempDir())
	cmd.Env = append(os.Environ(), "AUTOWONDER_TOKEN=fixture")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	var exitErr error
	go func() { exitErr = cmd.Wait(); close(done) }()
	defer func() { stop(); <-done }()
	connect := func() (*websocket.Conn, aw.Heartbeat) {
		select {
		case c := <-sockets:
			var h aw.Heartbeat
			if err := wsjson.Read(ctx, c, &h); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { c.CloseNow() })
			return c, h
		case <-done:
			t.Fatalf("executor exited: %v", exitErr)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		return nil, aw.Heartbeat{}
	}
	c, h := connect()
	prev, err := time.Parse(time.RFC3339Nano, h.StartedAt)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err = wsjson.Write(ctx, c, map[string]any{"type": "EXECUTOR_RESTART", "executorId": 7, "requestId": "apply", "launchConfigVersion": 7, "targetRuntimeVersion": "1.3.9", "previousStartedAt": prev.UnixMilli(), "restartMode": "WHEN_IDLE", "launch": map[string]any{"runtimeVersion": "1.3.9", "memoryMode": "none", "workspaceRoot": root, "maxConcurrentDispatches": 2}}); err != nil {
		t.Fatal(err)
	}
	var progress aw.ExecutorRestartResult
	if err = wsjson.Read(ctx, c, &progress); err != nil || progress.Status != "RESTARTING" {
		t.Fatalf("progress %v %v", progress, err)
	}
	c, h = connect()
	if h.RestartRequestID != "apply" || h.Version != "1.3.9" || h.MaxConcurrentDispatches != 2 {
		t.Fatalf("post-exec heartbeat %+v", h)
	}
	if err = wsjson.Write(ctx, c, map[string]any{"type": "CONVERSATION_TURN", "conversationId": 11, "turnId": 24, "content": "Reply only with the word lilac."}); err != nil {
		t.Fatal(err)
	}
	for {
		var reply struct{ Type, Status, ReplyMarkdown, Error string }
		if err = wsjson.Read(ctx, c, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Type == "CONVERSATION_TURN_ACK" {
			if reply.Status != "SUCCESS" || reply.ReplyMarkdown != "lilac" {
				t.Fatal(reply)
			}
			break
		}
	}
	items, err := os.ReadDir(root)
	if err != nil || len(items) != 1 {
		t.Fatalf("workspace %v %v", items, err)
	}
	if err = cmd.Process.Signal(os.Interrupt); err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Fatal(err)
	}
	select {
	case <-done:
		if exitErr != nil {
			t.Fatal(exitErr)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
