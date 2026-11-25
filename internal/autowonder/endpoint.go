package autowonder

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
)

// Endpoint wires application callbacks to Client.Run. Construct and register
// handlers before starting the client. Each executor identity owns its endpoint.
type Endpoint struct {
	snapshot func() Heartbeat
	pending  *Outbox
	handlers map[reflect.Type]func(context.Context, *Session, Incoming) error
}

// snapshot supplies a fresh inventory for each heartbeat. If an outbox is
// provided, pending dispatches are included until their terminal receipt arrives.
func NewEndpoint(snapshot func() Heartbeat, pending *Outbox) (*Endpoint, error) {
	if snapshot == nil {
		return nil, errors.New("heartbeat snapshot required")
	}
	return &Endpoint{snapshot: snapshot, pending: pending, handlers: make(map[reflect.Type]func(context.Context, *Session, Incoming) error)}, nil
}

// On binds one incoming frame type to a callback. Each type has one owner.
func On[T Incoming](e *Endpoint, f func(context.Context, *Session, T) error) error {
	if e == nil || f == nil {
		return errors.New("missing endpoint or frame handler")
	}
	typ := reflect.TypeFor[T]()
	if typ.Kind() != reflect.Pointer {
		return errors.New("register a concrete incoming message pointer")
	}
	if _, ok := e.handlers[typ]; ok {
		return fmt.Errorf("handler for %s already registered", typ)
	}
	e.handlers[typ] = func(ctx context.Context, s *Session, m Incoming) error { return f(ctx, s, m.(T)) }
	return nil
}
func (e *Endpoint) Heartbeat() Heartbeat {
	h := e.snapshot()
	if e.pending != nil && slices.Contains(h.ProtocolFeatures, FeatureDispatchInventoryV1) {
		// Never mutate a slice owned by the application or count an ID twice.
		seen := make(map[int64]bool, len(h.OwnedDispatchIDs))
		owned := make([]int64, 0, len(h.OwnedDispatchIDs))
		for _, id := range h.OwnedDispatchIDs {
			owned = append(owned, id)
			seen[id] = true
		}
		for _, id := range e.pending.OwnedDispatchIDs() {
			if !seen[id] {
				owned = append(owned, id)
				seen[id] = true
			}
		}
		if len(owned) > 1000 || h.DispatchInventoryError == "OWNED_DISPATCH_LIMIT_EXCEEDED" {
			ready := false
			h.DispatchInventoryReady = &ready
			h.DispatchInventoryError = "OWNED_DISPATCH_LIMIT_EXCEEDED"
			h.OwnedDispatchIDs = append([]int64{}, h.RunningDispatchIDs...)
		} else {
			h.OwnedDispatchIDs = owned
		}
	}
	return h
}
func (e *Endpoint) Connected(ctx context.Context, s *Session) error {
	if e.pending == nil {
		return nil
	}
	return e.pending.Replay(ctx, s)
}
func (e *Endpoint) Handle(ctx context.Context, s *Session, m Incoming) error {
	if e.pending != nil {
		if _, err := e.pending.HandleReceipt(ctx, s, m); err != nil {
			return err
		}
	}
	if f := e.handlers[reflect.TypeOf(m)]; f != nil {
		return f(ctx, s, m)
	}
	const unsupported = "operation not supported by this executor"
	switch v := m.(type) {
	case *TaskDispatch:
		return s.Send(ctx, &TaskBusy{DispatchID: v.DispatchID, Reason: "executor has no task handler"})
	case *TaskPause:
		return s.Send(ctx, &TaskPauseFailed{DispatchID: v.DispatchID, Error: unsupported})
	case *TaskGuidance:
		return s.Send(ctx, &TaskGuidanceAck{GuidanceID: v.GuidanceID, DispatchID: v.DispatchID, Status: "FAILED", Error: unsupported})
	case *TaskSteer:
		return s.Send(ctx, &TaskSteerAck{GuidanceID: v.GuidanceID, DispatchID: v.DispatchID, Status: "FAILED", Error: unsupported})
	case *QoderModelCatalogRequest:
		return s.Send(ctx, &QoderModelCatalogResult{Provider: v.Provider, RequestID: v.RequestID, Success: false})
	case *MCPConnectionTest:
		return s.Send(ctx, &MCPConnectionTestResult{TestID: v.TestID, Success: false, Message: unsupported})
	case *ExecutorRestart:
		return s.Send(ctx, &ExecutorRestartResult{RequestID: v.RequestID, Status: "FAILED", Message: unsupported})
	case *ExecutorUpgrade:
		return s.Send(ctx, &ExecutorUpgradeResult{RequestID: v.RequestID, Phase: "failed", Error: unsupported})
	case *ConversationCommandsProbe:
		return s.Send(ctx, &ConversationCommandsResult{ConversationID: v.ConversationID, Status: "FAILED", Error: unsupported})
	case *ConversationTurn:
		return e.Send(ctx, s, &ConversationTurnAck{ConversationID: v.ConversationID, TurnID: v.TurnID, DispatchAttempt: v.DispatchAttempt, ConversationDeliveryProtocolVersion: v.ConversationDeliveryProtocolVersion, Status: "FAILED", Error: unsupported})
	case *ConversationElicitationReply:
		if v.AnswerProtocolVersion == 1 {
			return s.Send(ctx, &ConversationElicitationReplyAck{ConversationID: v.ConversationID, TurnID: v.TurnID, DispatchAttempt: v.DispatchAttempt, AnswerProtocolVersion: 1, RequestID: v.RequestID, AnswerID: v.AnswerID, Status: "UNAVAILABLE"})
		}
		return nil
	case *Unknown, *TaskResultAck, *ConversationDeliveryAck, *ExecutorUpgradeResultAck, *ExecutorUpgradeResultSupport, *TaskHandoffResult:
		return nil
	default:
		return fmt.Errorf("no handler for %s", m.Kind())
	}
}

// Send records receipt-backed deliveries before writing them. A failed network
// write leaves the original payload available for replay. Pass a nil session to
// queue a receipt-backed delivery while offline.
func (e *Endpoint) Send(ctx context.Context, s *Session, m Outgoing) error {
	if m == nil || reflect.ValueOf(m).IsNil() {
		return errors.New("missing outgoing message")
	}
	durable := false
	switch msg := m.(type) {
	case *TaskResult:
		durable = msg.ResultProtocolVersion == 2
	case *ConversationTurnEvent:
		durable = msg.ConversationDeliveryProtocolVersion == 1
	case *ConversationTurnAck:
		durable = msg.ConversationDeliveryProtocolVersion == 1
	case *ExecutorUpgradeResult:
		durable = msg.ReceiptID != "" && msg.Phase != "receipt_probe"
	}
	if durable {
		if e.pending == nil {
			return errors.New("receipt-backed delivery requires an outbox")
		}
		if err := e.pending.Store(m); err != nil {
			return err
		}
	}
	if s == nil {
		if durable {
			return nil
		}
		return errors.New("outgoing message requires a connected session")
	}
	return s.Send(ctx, m)
}
