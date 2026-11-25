package autowonder

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestEndpointUpgradeReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer s.CloseNow()
		read := func() (string, string) {
			_, b, err := s.Read(ctx)
			if err != nil {
				t.Error(err)
				return "", ""
			}
			var f struct{ Type, Phase string }
			if err = json.Unmarshal(b, &f); err != nil {
				t.Error(err)
			}
			return f.Type, f.Phase
		}
		if typ, _ := read(); typ != "HEARTBEAT" {
			t.Error("missing heartbeat")
			return
		}
		if typ, phase := read(); typ != "EXECUTOR_UPGRADE_RESULT" || phase != "receipt_probe" {
			t.Error("sent terminal result before probe")
			return
		}
		if err := s.Write(ctx, websocket.MessageText, []byte(`{"type":"EXECUTOR_UPGRADE_RESULT_SUPPORT","requestId":"u","receiptId":"r"}`)); err != nil {
			t.Error(err)
			return
		}
		if typ, phase := read(); typ != "EXECUTOR_UPGRADE_RESULT" || phase != "success" {
			t.Error("missing terminal result")
			return
		}
		if err := s.Write(ctx, websocket.MessageText, []byte(`{"type":"EXECUTOR_UPGRADE_RESULT_ACK","requestId":"u","receiptId":"r"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ExecutorID: 42, Token: "test"})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	q, err := c.OpenOutbox(dir)
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEndpoint(func() Heartbeat { return Heartbeat{} }, q)
	if err != nil {
		t.Fatal(err)
	}
	if err = e.Send(ctx, nil, &ExecutorUpgradeResult{RequestID: "u", ReceiptID: "r", Phase: "success", CurrentVersion: "1.3.0"}); err != nil {
		t.Fatal(err)
	}
	stop := errors.New("receipt received")
	if err := On(e, func(context.Context, *Session, *ExecutorUpgradeResultAck) error { return stop }); err != nil {
		t.Fatal(err)
	}
	if err = c.Run(ctx, time.Minute, e); !errors.Is(err, stop) {
		t.Fatal(err)
	}
	reopened, err := c.OpenOutbox(dir)
	if err != nil || len(reopened.pending) != 0 {
		t.Fatal("result not cleared", err)
	}
}

func TestDebugLogAndDiagnostics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/signed":
			if r.Method != "PUT" || r.Header.Get("Authorization") != "" || r.URL.Query().Has("token") {
				t.Error("unexpected signed upload credentials")
			}
			b, err := io.ReadAll(r.Body)
			if err != nil || string(b) != "test" {
				t.Error("invalid log bytes")
			}
			w.WriteHeader(204)
		case "/api/daemon/dispatches/71/diagnostics":
			if r.Method != "POST" || r.URL.Query().Get("token") != "t" {
				t.Error("wrong diagnostics route")
			}
			f, _, err := r.FormFile("files")
			if err != nil {
				t.Error(err)
				return
			}
			f.Close()
			io.WriteString(w, `{"files":[{"index":0,"status":"ACCEPTED","path":"test.json"}]}`)
		default:
			t.Error("unknown request")
		}
	}))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ExecutorID: 42, Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if err = c.PutDebugLog(t.Context(), DebugLogUploadLink{UploadURL: srv.URL + "/signed"}, strings.NewReader("test"), 4); err != nil {
		t.Fatal(err)
	}
	if _, err = c.UploadDiagnostics(t.Context(), 71, []UploadFile{{"test.json", strings.NewReader("{}")}}, "", nil); err != nil {
		t.Fatal(err)
	}
}
func TestRoutingAndRetryPolicy(t *testing.T) {
	if _, err := decodeForExecutor([]byte(`{"type":"TASK_DISPATCH","executorId":7}`), 42); err == nil {
		t.Fatal("accepted another executor's dispatch")
	}
	if shouldReconnect(websocket.CloseError{Code: 4001}) || shouldReconnect(websocket.CloseError{Code: 1008, Reason: "authentication failed"}) || shouldReconnect(&HTTPError{StatusCode: 403}) {
		t.Fatal("would reconnect on terminal error")
	}
	if !shouldReconnect(websocket.CloseError{Code: 1008, Reason: "session initialization busy"}) || !shouldReconnect(&HTTPError{StatusCode: 503}) {
		t.Fatal("transient failure treated as permanent")
	}
	e, _ := NewEndpoint(func() Heartbeat { return Heartbeat{} }, nil)
	cb := func(context.Context, *Session, *TaskDispatch) error { return nil }
	if err := On(e, cb); err != nil {
		t.Fatal(err)
	}
	if err := On(e, cb); err == nil {
		t.Fatal("replaced handler")
	}
}
