package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestUpgradeVersionSurvivesRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	conns := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		conns <- c
	}))
	defer srv.Close()
	dir := t.TempDir()
	start := func(id int64, version string) (*websocket.Conn, func()) {
		t.Helper()
		run, stop := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() {
			done <- Run(run, aw.Config{ExecutorID: id, Token: "test", BaseURL: srv.URL}, dir, &stub{}, nil)
		}()
		select {
		case c := <-conns:
			var h aw.Heartbeat
			if err := wsjson.Read(ctx, c, &h); err != nil || h.Version != version || !slices.Contains(h.ProtocolFeatures, aw.FeatureExecutorUpdateV1) {
				stop()
				t.Fatalf("heartbeat %v: %v", h, err)
			}
			return c, func() { stop(); c.CloseNow(); <-done }
		case err := <-done:
			stop()
			t.Fatal(err)
		case <-ctx.Done():
			stop()
			t.Fatal(ctx.Err())
		}
		return nil, nil
	}
	c, stop := start(1, "aws-codex-test")
	upgrade := func(id, version string) aw.ExecutorUpgradeResult {
		t.Helper()
		if err := wsjson.Write(ctx, c, map[string]any{"type": "EXECUTOR_UPGRADE", "requestId": id, "targetVersion": version, "resultReceiptAck": true}); err != nil {
			t.Fatal(err)
		}
		var result aw.ExecutorUpgradeResult
		if err := wsjson.Read(ctx, c, &result); err != nil || result.RequestID != id || result.Phase != "success" || result.CurrentVersion != version || result.ReceiptID == "" {
			t.Fatalf("result %v: %v", result, err)
		}
		if err := wsjson.Write(ctx, c, map[string]string{"type": "EXECUTOR_UPGRADE_RESULT_ACK", "receiptId": result.ReceiptID, "requestId": id}); err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := upgrade("A", "1.3.1")
	upgrade("B", "1.3.2")
	if replay := upgrade("A", "1.3.1"); replay.ReceiptID != first.ReceiptID {
		t.Fatal(replay)
	}
	stop()
	_, stop = start(1, "1.3.2")
	stop()
	_, stop = start(2, "aws-codex-test")
	stop()
}
