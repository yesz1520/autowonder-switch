package executor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"unicode/utf8"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

type eventStream struct {
	mu  sync.Mutex
	seq int64
	r   *runtime
	q   *aw.ConversationTurn
}

func (e *eventStream) emit(kind string, value any) {
	b, _ := json.Marshal(value)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.seq++
	var chunks []string
	for len(b) > 0 {
		n := min(16<<10, len(b))
		for n < len(b) && !utf8.RuneStart(b[n]) {
			n--
		}
		chunks = append(chunks, string(b[:n]))
		b = b[n:]
	}
	for i, v := range chunks {
		e.r.report(&aw.ConversationTurnEvent{ConversationID: e.q.ConversationID, TurnID: e.q.TurnID, DispatchAttempt: e.q.DispatchAttempt, ConversationDeliveryProtocolVersion: e.q.ConversationDeliveryProtocolVersion, EventSeq: e.seq, ChunkIndex: int64(i), ChunkCount: int64(len(chunks)), EventType: kind, PayloadFragment: v})
	}
}

type pendingQuestion struct {
	q      *aw.ConversationTurn
	answer chan *aw.ConversationElicitationReply
	id     string
}

func (e *eventStream) ask(ctx context.Context, msg string, schema map[string]any) (harness.Answer, error) {
	id := aw.NewResultID()
	p := &pendingQuestion{q: e.q, answer: make(chan *aw.ConversationElicitationReply, 1)}
	r := e.r
	r.mu.Lock()
	if r.questions == nil {
		r.questions = map[string]*pendingQuestion{}
	}
	r.questions[id] = p
	r.mu.Unlock()
	e.emit("acp_elicitation", map[string]any{"type": "acp_elicitation", "data": map[string]any{"requestId": id, "mode": "form", "message": msg, "requestedSchema": schema, "answerProtocolVersion": e.q.ConversationAnswerProtocolVersion}})
	select {
	case <-ctx.Done():
		r.mu.Lock()
		delete(r.questions, id)
		r.mu.Unlock()
		return harness.Answer{}, ctx.Err()
	case a := <-p.answer:
		var content map[string]any
		_ = json.Unmarshal(a.Content, &content)
		return harness.Answer{Action: a.Action, Content: content, Written: func(ok bool) {
			status := "UNKNOWN"
			if ok {
				status = "WRITTEN"
			}
			if a.AnswerProtocolVersion == 1 {
				if err := save(r.answerPath(a), status); err != nil {
					status = "UNKNOWN"
				}
				r.report(answerAck(a, status))
			}
			if ok {
				e.emit("acp_elicitation_resolved", map[string]any{"type": "acp_elicitation_resolved", "data": map[string]string{"requestId": id, "action": a.Action}})
			}
			r.mu.Lock()
			delete(r.questions, id)
			r.mu.Unlock()
		}}, nil
	}
}
func answerAck(a *aw.ConversationElicitationReply, status string) *aw.ConversationElicitationReplyAck {
	return &aw.ConversationElicitationReplyAck{ConversationID: a.ConversationID, TurnID: a.TurnID, DispatchAttempt: a.DispatchAttempt, AnswerProtocolVersion: 1, RequestID: a.RequestID, AnswerID: a.AnswerID, Status: status}
}
func (r *runtime) answerPath(a *aw.ConversationElicitationReply) string {
	return filepath.Join(r.dir, fmt.Sprintf("answer-%x", sha256.Sum256([]byte(fmt.Sprintf("%d/%d/%d/%q/%q", a.ConversationID, a.TurnID, a.DispatchAttempt, a.RequestID, a.AnswerID)))))
}
func (r *runtime) answer(ctx context.Context, s *aw.Session, a *aw.ConversationElicitationReply) error {
	reliable := a.AnswerProtocolVersion == 1
	if reliable {
		var status string
		if ok, err := cached(r.answerPath(a), &status); err != nil {
			return err
		} else if ok {
			if status == "UNKNOWN" {
				r.mu.Lock()
				p := r.questions[a.RequestID]
				if p != nil && p.id == a.AnswerID {
					status = "RECEIVED"
				}
				r.mu.Unlock()
			}
			return s.Send(ctx, answerAck(a, status))
		}
	}
	r.mu.Lock()
	p := r.questions[a.RequestID]
	status := "UNAVAILABLE"
	if p != nil && p.q.ConversationID == a.ConversationID && p.q.TurnID == a.TurnID && (!reliable || p.q.DispatchAttempt == a.DispatchAttempt && p.q.ConversationAnswerProtocolVersion == 1 && a.AnswerID != "") {
		if p.id != "" && p.id == a.AnswerID {
			status = "RECEIVED"
		} else if p.id == "" && (!reliable || (a.ReceiptOnly == nil || !*a.ReceiptOnly) && a.RemainingTTLMS > 0) && (a.Action == "accept" || a.Action == "decline" || a.Action == "cancel") {
			p.id = a.AnswerID
			if !reliable {
				p.id = "legacy"
			}
			status = "RECEIVED"
			// Order RECEIVED before delivery to the native RPC writer.
			if reliable {
				if err := save(r.answerPath(a), "UNKNOWN"); err != nil {
					p.id = ""
					r.mu.Unlock()
					return err
				}
				// Once accepted, deliver even if the receipt socket closes.
				// A retry can recover the final status from disk.
				r.mu.Unlock()
				err := s.Send(ctx, answerAck(a, status))
				p.answer <- a
				return err
			}
			p.answer <- a
			r.mu.Unlock()
			return nil
		}
	}
	r.mu.Unlock()
	if reliable {
		return s.Send(ctx, answerAck(a, status))
	}
	return nil
}
