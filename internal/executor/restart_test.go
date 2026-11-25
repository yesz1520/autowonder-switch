package executor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestRestartHandshake(t *testing.T) {
	for _, apply := range []bool{false, true} {
		t.Run(fmt.Sprintf("launch-%t", apply), func(t *testing.T) { testRestartHandshake(t, apply) })
	}
}
func testRestartHandshake(t *testing.T, apply bool) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	capacities := make(chan string, 1)
	conns := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, e := websocket.Accept(w, r, nil)
		if e == nil {
			capacities <- r.URL.Query().Get("maxConcurrentDispatches")
			conns <- c
		}
	}))
	defer srv.Close()
	dir := t.TempDir()
	start := func() (<-chan error, *websocket.Conn, aw.Heartbeat) {
		done := make(chan error, 1)
		go func() {
			done <- Run(ctx, aw.Config{ExecutorID: 1, Token: "test", BaseURL: srv.URL}, dir, &stub{}, nil, FileOptions{RestartEnabled: true})
		}()
		select {
		case c := <-conns:
			var h aw.Heartbeat
			if err := wsjson.Read(ctx, c, &h); err != nil || !slices.Contains(h.ProtocolFeatures, aw.FeatureExecutorRestartV1) {
				t.Fatalf("heartbeat %+v: %v", h, err)
			}
			return done, c, h
		case e := <-done:
			t.Fatalf("early exit %v", e)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		return nil, nil, aw.Heartbeat{}
	}
	done, c, h := start()
	defer c.CloseNow()
	if capacity := <-capacities; capacity != "1" {
		t.Fatalf("initial handshake capacity %s", capacity)
	}
	prev, e := time.Parse(time.RFC3339Nano, h.StartedAt)
	if e != nil {
		t.Fatal(e)
	}
	req := map[string]any{"type": "EXECUTOR_RESTART", "requestId": "one", "restartMode": "WHEN_IDLE", "previousStartedAt": prev.UnixMilli()}
	root := t.TempDir()
	if apply {
		if !slices.Contains(h.ProtocolFeatures, aw.FeatureExecutorLaunchCommandV1) {
			t.Fatal(h.ProtocolFeatures)
		}
		req["launchConfigVersion"] = 4
		req["targetRuntimeVersion"] = "1.3.7"
		req["launch"] = map[string]any{"runtimeVersion": "1.3.7", "workspaceRoot": root, "maxConcurrentDispatches": 3, "debug": true, "daemonName": "runner", "memoryMode": "none"}
	}
	if e = wsjson.Write(ctx, c, req); e != nil {
		t.Fatal(e)
	}
	var res aw.ExecutorRestartResult
	if e = wsjson.Read(ctx, c, &res); e != nil || res.Status != "RESTARTING" {
		t.Fatalf("%+v %v", res, e)
	}
	select {
	case e = <-done:
		if !errors.Is(e, ErrRestartRequested) {
			t.Fatal(e)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	done, c, h = start()
	defer c.CloseNow()
	capacity := <-capacities
	if apply {
		if h.Version != "1.3.7" || h.MaxConcurrentDispatches != 3 || capacity != "3" {
			t.Fatalf("applied config %+v capacity=%s", h, capacity)
		}
		if e = wsjson.Write(ctx, c, map[string]any{"type": "CONVERSATION_TURN", "conversationId": 9, "turnId": 881, "dispatchAttempt": 1, "content": "hello"}); e != nil {
			t.Fatal(e)
		}
		for {
			var m struct{ Type, Status string }
			if e = wsjson.Read(ctx, c, &m); e != nil {
				t.Fatal(e)
			}
			if m.Type == "CONVERSATION_TURN_ACK" {
				if m.Status != "SUCCESS" {
					t.Fatal(m)
				}
				break
			}
		}
		children, e := os.ReadDir(root)
		if e != nil || len(children) != 1 || !children[0].IsDir() {
			t.Fatalf("workspace %v %v", children, e)
		}
		links, e := filepath.Glob(filepath.Join(dir, "*", "conversation-9", "workspace"))
		if e != nil || len(links) != 1 {
			t.Fatalf("links %v %v", links, e)
		}
		target, e := os.Readlink(links[0])
		if e != nil || target != filepath.Join(root, children[0].Name()) {
			t.Fatalf("target %s %v", target, e)
		}
	}
	next, e := time.Parse(time.RFC3339Nano, h.StartedAt)
	if e != nil || next.UnixMilli() <= prev.UnixMilli() || h.RestartRequestID != "one" {
		t.Fatalf("restart heartbeat %+v %v", h, e)
	}
	if e = wsjson.Write(ctx, c, req); e != nil {
		t.Fatal(e)
	}
	req["requestId"] = "two"
	req["restartMode"] = "invalid"
	if e = wsjson.Write(ctx, c, req); e != nil {
		t.Fatal(e)
	}
	if e = wsjson.Read(ctx, c, &res); e != nil || res.Status != "FAILED" || res.RequestID != "two" {
		t.Fatalf("%+v %v", res, e)
	}
	cancel()
	<-done
}

func TestRestartDuringActiveTurn(t *testing.T) {
	for _, mode := range []string{"WHEN_IDLE", "FORCE"} {
		t.Run(mode, func(t *testing.T) {
			ctx, stop := context.WithTimeout(t.Context(), 8*time.Second)
			defer stop()
			conns := make(chan *websocket.Conn, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, e := websocket.Accept(w, r, nil)
				if e == nil {
					conns <- c
				}
			}))
			defer srv.Close()
			b := &waiting{started: make(chan struct{}, 1)}
			dir := t.TempDir()
			done := make(chan error, 1)
			go func() {
				done <- Run(ctx, aw.Config{ExecutorID: 1, Capacity: 2, BaseURL: srv.URL, Token: "test"}, dir, b, nil, FileOptions{RestartEnabled: true})
			}()
			var c *websocket.Conn
			select {
			case c = <-conns:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case e := <-done:
				t.Fatal(e)
			}
			defer c.CloseNow()
			read := func(kind string) string {
				t.Helper()
				for {
					var m struct{ Type, Status string }
					if err := wsjson.Read(ctx, c, &m); err != nil {
						t.Fatal(err)
					}
					if m.Type == kind {
						return m.Status
					}
				}
			}
			send := func(m map[string]any) {
				t.Helper()
				if err := wsjson.Write(ctx, c, m); err != nil {
					t.Fatal(err)
				}
			}
			read("HEARTBEAT")
			send(map[string]any{"type": "CONVERSATION_TURN", "conversationId": 1, "turnId": 1, "dispatchAttempt": 1, "content": "wait"})
			select {
			case <-b.started:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			send(map[string]any{"type": "EXECUTOR_RESTART", "requestId": "reboot", "restartMode": mode})
			if mode == "WHEN_IDLE" {
				if got := read("EXECUTOR_RESTART_RESULT"); got != "WAITING_IDLE" {
					t.Fatal(got)
				}
				send(map[string]any{"type": "CONVERSATION_TURN", "conversationId": 2, "turnId": 2, "dispatchAttempt": 1, "content": "new work"})
				if got := read("CONVERSATION_TURN_ACK"); got != "FAILED" {
					t.Fatal(got)
				}
				send(map[string]any{"type": "CONVERSATION_TURN_CANCEL", "conversationId": 1, "turnId": 1, "dispatchAttempt": 1})
				if got := read("CONVERSATION_TURN_ACK"); got != "CANCELED" {
					t.Fatal(got)
				}
			}
			if got := read("EXECUTOR_RESTART_RESULT"); got != "RESTARTING" {
				t.Fatal(got)
			}
			select {
			case err := <-done:
				if !errors.Is(err, ErrRestartRequested) {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
	}
}
