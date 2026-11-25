package autowonder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestOutboxReopenAndScope(t *testing.T) {
	dir := t.TempDir()
	first, err := New(Config{BaseURL: "https://platform.example", ExecutorID: 42, Token: "one"})
	if err != nil {
		t.Fatal(err)
	}
	queue, err := first.OpenOutbox(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := &TaskResult{DispatchID: 71, ResultProtocolVersion: 2, ResultID: NewResultID(), Error: "failed"}
	if err = queue.Store(result); err != nil {
		t.Fatal(err)
	}
	if err = queue.Store(result); err != nil {
		t.Fatal(err)
	}
	result.Error = "different"
	if err = queue.Store(result); err == nil {
		t.Fatal("overwrote existing result")
	}
	rotated, _ := New(Config{BaseURL: "https://platform.example", ExecutorID: 42, Token: "two"})
	queue, err = rotated.OpenOutbox(dir)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := New(Config{BaseURL: "https://platform.example", ExecutorID: 43, Token: "three"})
	q2, err := other.OpenOutbox(dir)
	if err != nil || len(q2.OwnedDispatchIDs()) != 0 {
		t.Fatal("identity leaked", err)
	}
	ack := &TaskResultAck{ResultProtocolVersion: 2, DispatchID: 71, ResultID: result.ResultID, Disposition: "RETRYABLE"}
	if removed, err := queue.Confirm(ack); err != nil || removed || !slices.Equal(queue.OwnedDispatchIDs(), []int64{71}) {
		t.Fatal("lost retryable result", err)
	}
	ack.Disposition = "REJECTED"
	if removed, err := queue.Confirm(ack); err != nil || !removed {
		t.Fatal("did not remove rejected result", err)
	}
	event := &ConversationTurnEvent{ConversationDeliveryProtocolVersion: 1, ConversationID: 1, TurnID: 2, DispatchAttempt: 1, EventSeq: 5, ChunkIndex: 0, ChunkCount: 2, PayloadFragment: "part"}
	if err = queue.Store(event); err != nil {
		t.Fatal(err)
	}
	receipt := &ConversationDeliveryAck{ConversationDeliveryProtocolVersion: 1, FrameType: event.Kind(), ConversationID: 1, TurnID: 2, DispatchAttempt: 1, EventSeq: 5, ChunkIndex: 1}
	if removed, err := queue.Confirm(receipt); err != nil || removed {
		t.Fatal("matched another chunk", err)
	}
	receipt.ChunkIndex = 0
	if removed, err := queue.Confirm(receipt); err != nil || !removed {
		t.Fatal("lost chunk receipt", err)
	}
}

type resultPeer struct {
	queue  *Outbox
	stop   error
	result *TaskResult
}

func (p *resultPeer) Heartbeat() Heartbeat                            { return Heartbeat{Version: "test"} }
func (p *resultPeer) Connected(ctx context.Context, s *Session) error { return p.queue.Replay(ctx, s) }
func (p *resultPeer) Handle(ctx context.Context, s *Session, message Incoming) error {
	switch m := message.(type) {
	case *TaskDispatch:
		if err := s.Send(ctx, &TaskAck{DispatchID: m.DispatchID}); err != nil {
			return err
		}
		if err := p.queue.Store(p.result); err != nil {
			return err
		}
		return s.Send(ctx, p.result)
	case *TaskResultAck:
		confirmed, err := p.queue.Confirm(m)
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("unexpected result receipt")
		}
		return p.stop
	default:
		return fmt.Errorf("unexpected %s", message.Kind())
	}
}
func TestReconnectReplaysBeforeDispatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	resultID := NewResultID()
	var connections atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer s.CloseNow()
		read := func(want string) bool {
			_, p, err := s.Read(ctx)
			if err != nil {
				t.Error(err)
				return false
			}
			var f struct{ Type, ResultID string }
			if err = json.Unmarshal(p, &f); err != nil || f.Type != want || (want == "TASK_RESULT" && f.ResultID != resultID) {
				t.Errorf("frame %s: %s, %v", want, p, err)
				return false
			}
			return true
		}
		if !read("HEARTBEAT") {
			return
		}
		if connections.Add(1) == 1 {
			if err := s.Write(ctx, websocket.MessageText, []byte(`{"type":"TASK_DISPATCH","executorId":42,"dispatchId":71}`)); err != nil {
				t.Error(err)
				return
			}
			if !read("TASK_ACK") || !read("TASK_RESULT") {
				return
			}
			return // socket loss before the receipt
		}
		if !read("TASK_RESULT") {
			return
		}
		receipt := fmt.Sprintf(`{"type":"TASK_RESULT_ACK","dispatchId":71,"resultProtocolVersion":2,"resultId":%q,"accepted":true,"disposition":"ACCEPTED"}`, resultID)
		if err = s.Write(ctx, websocket.MessageText, []byte(receipt)); err != nil {
			t.Error(err)
			return
		}
	}))
	defer server.Close()
	c, err := New(Config{BaseURL: server.URL, ExecutorID: 42, Token: "test"})
	if err != nil {
		t.Fatal(err)
	}
	q, err := c.OpenOutbox(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stop := errors.New("received result acknowledgement")
	h := &resultPeer{q, stop, &TaskResult{DispatchID: 71, ResultID: resultID, ResultProtocolVersion: 2, Error: "no runner configured"}}
	if err = c.Run(ctx, time.Minute, h); !errors.Is(err, stop) {
		t.Fatal(err)
	}
	if connections.Load() != 2 || len(q.OwnedDispatchIDs()) != 0 {
		t.Fatal("did not finish delivery")
	}
}
