package autowonder

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode/utf16"
)

var resultIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func NewResultID() string {
	var id [16]byte
	rand.Read(id[:])
	id[6] = (id[6] & 15) | 64
	id[8] = (id[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}

type deliveryKey struct {
	Kind                                        string
	RequestID, ReceiptID                        string
	DispatchID                                  int64
	ResultID                                    string
	ConversationID, TurnID, Attempt, Seq, Chunk int64
}
type delivery struct {
	key   deliveryKey
	frame json.RawMessage
}

// Outbox keeps V2 task results and V1 conversation deliveries and terminal upgrade receipts across disconnects
// and process restarts. Open once per identity in the owning daemon. Store before
// sending, replay in Connected, and call Confirm on incoming receipts.
type Outbox struct {
	mu      sync.Mutex
	path    string
	pending []delivery
}

func (c *Client) OpenOutbox(dir string) (*Outbox, error) {
	if dir == "" {
		return nil, errors.New("missing outbox directory")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	identity := fmt.Sprintf("%s/%d", c.base.Redacted(), c.id)
	path := filepath.Join(dir, fmt.Sprintf("%x.json", sha256.Sum256([]byte(identity))))
	q := &Outbox{path: path}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return q, nil
	}
	if err != nil {
		return nil, err
	}
	var frames []json.RawMessage
	if err = json.Unmarshal(b, &frames); err != nil {
		return nil, err
	}
	seen := map[deliveryKey]bool{}
	for _, f := range frames {
		k, err := keyForDelivery(f)
		if err != nil {
			return nil, err
		}
		if seen[k] {
			return nil, errors.New("duplicate persisted delivery")
		}
		seen[k] = true
		q.pending = append(q.pending, delivery{k, f})
	}
	return q, nil
}
func keyForDelivery(b []byte) (deliveryKey, error) {
	var v struct {
		RequestID, ReceiptID, Phase                                                                        string
		Terminal                                                                                           bool
		Type                                                                                               string
		DispatchID, ResultProtocolVersion                                                                  int64
		ResultID                                                                                           string
		ConversationDeliveryProtocolVersion, ConversationID, TurnID, DispatchAttempt, EventSeq, ChunkIndex int64
		Success                                                                                            bool
		CheckpointSeq                                                                                      int64
		CheckpointSHA256                                                                                   string
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return deliveryKey{}, err
	}
	switch v.Type {
	case "TASK_RESULT":
		if v.DispatchID <= 0 || v.ResultProtocolVersion != 2 || !resultIDPattern.MatchString(v.ResultID) {
			return deliveryKey{}, errors.New("invalid V2 result identity")
		}
		if v.Success && (v.CheckpointSeq <= 0 || v.CheckpointSHA256 == "") {
			return deliveryKey{}, errors.New("success requires a completion checkpoint")
		}
		return deliveryKey{Kind: v.Type, DispatchID: v.DispatchID, ResultID: v.ResultID}, nil
	case "EXECUTOR_UPGRADE_RESULT":
		if v.RequestID == "" || strings.TrimSpace(v.ReceiptID) == "" || len(utf16.Encode([]rune(v.ReceiptID))) > 128 || !(v.Phase == "success" || v.Phase == "failed" && v.Terminal) {
			return deliveryKey{}, errors.New("expected terminal upgrade result and receipt ID")
		}
		return deliveryKey{Kind: v.Type, RequestID: v.RequestID, ReceiptID: v.ReceiptID}, nil
	case "CONVERSATION_TURN_EVENT", "CONVERSATION_TURN_ACK":
		if v.ConversationDeliveryProtocolVersion != 1 || v.ConversationID <= 0 || v.TurnID <= 0 {
			return deliveryKey{}, errors.New("expected a V1 turn delivery")
		}
		return deliveryKey{Kind: v.Type, ConversationID: v.ConversationID, TurnID: v.TurnID, Attempt: v.DispatchAttempt, Seq: v.EventSeq, Chunk: v.ChunkIndex}, nil
	default:
		return deliveryKey{}, errors.New("frame does not use task or conversation receipts")
	}
}
func (q *Outbox) Store(msg Outgoing) error {
	b, err := Encode(msg)
	if err != nil {
		return err
	}
	key, err := keyForDelivery(b)
	if err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, entry := range q.pending {
		if entry.key == key {
			if !bytes.Equal(entry.frame, b) {
				return errors.New("delivery identity already has a different payload")
			}
			return nil
		}
	}
	return q.replace(append(slices.Clone(q.pending), delivery{key, b}))
}

func (q *Outbox) replace(entries []delivery) error {
	dir := filepath.Dir(q.path)
	tmp, err := os.CreateTemp(dir, ".outbox-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	frames := make([]json.RawMessage, 0, len(entries))
	for _, item := range entries {
		frames = append(frames, item.frame)
	}
	err = json.NewEncoder(tmp).Encode(frames)
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(tmp.Name(), q.path); err != nil {
		return err
	}
	q.pending = entries
	folder, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer folder.Close()
	return folder.Sync()
}

// Replay resends immutable payloads in enqueue order. Concurrent confirmations
// may result in a duplicate transmission, which retains the original identity.
func (q *Outbox) Replay(ctx context.Context, s *Session) error {
	q.mu.Lock()
	frames := slices.Clone(q.pending)
	q.mu.Unlock()
	for _, entry := range frames {
		if entry.key.Kind == "EXECUTOR_UPGRADE_RESULT" {
			err := s.Send(ctx, &ExecutorUpgradeResult{RequestID: entry.key.RequestID, ReceiptID: entry.key.ReceiptID, Phase: "receipt_probe"})
			if err != nil {
				return err
			}
			continue
		}
		if err := s.write(ctx, entry.frame); err != nil {
			return err
		}
	}
	return nil
}

// Confirm removes a matching terminal receipt. Retryable and unrelated receipts
// leave the delivery queued. A terminal rejection is still visible to the caller.
func (q *Outbox) Confirm(msg Incoming) (bool, error) {
	var k deliveryKey
	switch a := msg.(type) {
	case *ExecutorUpgradeResultAck:
		k = deliveryKey{Kind: "EXECUTOR_UPGRADE_RESULT", RequestID: a.RequestID, ReceiptID: a.ReceiptID}
	case *TaskResultAck:
		if a.ResultProtocolVersion != 2 || a.ResultID == "" || !(a.Disposition == "ACCEPTED" && a.Accepted || a.Disposition == "REJECTED" && !a.Accepted) {
			return false, nil
		}
		k = deliveryKey{Kind: "TASK_RESULT", DispatchID: a.DispatchID, ResultID: a.ResultID}
	case *ConversationDeliveryAck:
		if a.ConversationDeliveryProtocolVersion != 1 || (a.FrameType != "CONVERSATION_TURN_ACK" && a.FrameType != "CONVERSATION_TURN_EVENT") {
			return false, nil
		}
		k = deliveryKey{Kind: a.FrameType, ConversationID: a.ConversationID, TurnID: a.TurnID, Attempt: a.DispatchAttempt, Seq: a.EventSeq, Chunk: a.ChunkIndex}
	default:
		return false, nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, e := range q.pending {
		if e.key == k {
			err := q.replace(slices.Delete(slices.Clone(q.pending), i, i+1))
			return err == nil, err
		}
	}
	return false, nil
}
func (q *Outbox) OwnedDispatchIDs() []int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	ids := []int64{}
	seen := map[int64]bool{}
	for _, e := range q.pending {
		if id := e.key.DispatchID; id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// HandleReceipt completes terminal deliveries. An upgrade support frame only
// authorizes sending the result on this connection; it never removes the result.
func (q *Outbox) HandleReceipt(ctx context.Context, s *Session, msg Incoming) (bool, error) {
	support, ok := msg.(*ExecutorUpgradeResultSupport)
	if !ok {
		return q.Confirm(msg)
	}
	key := deliveryKey{Kind: "EXECUTOR_UPGRADE_RESULT", RequestID: support.RequestID, ReceiptID: support.ReceiptID}
	q.mu.Lock()
	var frame json.RawMessage
	for _, e := range q.pending {
		if e.key == key {
			frame = e.frame
			break
		}
	}
	q.mu.Unlock()
	if frame == nil {
		return false, nil
	}
	return false, s.write(ctx, frame)
}
