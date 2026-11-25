package executor

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestStartupRecoversFinalDeliveries(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	conns := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err == nil {
			conns <- c
		}
	}))
	defer srv.Close()
	dir := filepath.Join(t.TempDir(), "executor [state]")
	cfg := aw.Config{ExecutorID: 7, Token: "test", BaseURL: srv.URL}
	root := filepath.Join(dir, fmt.Sprintf("%x", sha256.Sum256([]byte(srv.URL+"/7"))))
	task := &aw.TaskResult{DispatchID: 41, ResultProtocolVersion: 2, ResultID: aw.NewResultID(), Error: "archived failure"}
	turn := &aw.ConversationTurnAck{ConversationID: 9, TurnID: 8, DispatchAttempt: 1, ConversationDeliveryProtocolVersion: 1, Status: "SUCCESS", ReplyMarkdown: "saved answer"}
	for path, msg := range map[string]aw.Outgoing{"task-41/result.json": task, "conversation-9/turn-8-1.json": turn} {
		dest := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			t.Fatal(err)
		}
		if err := save(dest+".ready", msg); err != nil {
			t.Fatal(err)
		}
	}
	// Simulate a crash after enqueue but before rename for one of the two frames.
	cl, err := aw.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	box, err := cl.OpenOutbox(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = box.Store(task); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, dir, &stub{}, nil) }()
	var c *websocket.Conn
	select {
	case c = <-conns:
	case err = <-done:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	defer c.CloseNow()
	var h aw.Heartbeat
	if err = wsjson.Read(ctx, c, &h); err != nil || !slices.Contains(h.OwnedDispatchIDs, 41) {
		t.Fatalf("inventory %+v %v", h, err)
	}
	var got aw.TaskResult
	if err = wsjson.Read(ctx, c, &got); err != nil || got.ResultID != task.ResultID || got.Error != task.Error {
		t.Fatalf("task %+v %v", got, err)
	}
	var ack aw.ConversationTurnAck
	if err = wsjson.Read(ctx, c, &ack); err != nil || !reflect.DeepEqual(ack, *turn) {
		t.Fatalf("turn %+v %v", ack, err)
	}
	for _, path := range []string{"task-41/result.json", "conversation-9/turn-8-1.json"} {
		if _, err = os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
		if _, err = os.Stat(filepath.Join(root, path+".ready")); !os.IsNotExist(err) {
			t.Fatal("pending journal remained", err)
		}
	}
	cancel()
	<-done
}
