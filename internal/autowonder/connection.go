package autowonder

import (
	"context"
	"errors"
	"time"

	"github.com/coder/websocket"
)

// Handler receives a fresh session on every connection. Connected runs after
// the first heartbeat and before incoming dispatches are handled. It can replay
// the identity's pending deliveries. Run execution contexts belong to the
// executor, not to this connection.
type Handler interface {
	Heartbeat() Heartbeat
	Connected(context.Context, *Session) error
	Handle(context.Context, *Session, Incoming) error
}
type handlerError struct{ cause error }

func (e *handlerError) Error() string { return e.cause.Error() }
func (e *handlerError) Unwrap() error { return e.cause }

// Run reconnects after a transport interruption. An invalid credential, a
// replacement by another session, and application errors return to the owner.
func (c *Client) Run(ctx context.Context, interval time.Duration, h Handler) error {
	if h == nil || interval <= 0 || interval >= 90*time.Second {
		return errors.New("invalid handler or heartbeat interval")
	}
	backoff := time.Second
	for {
		started := time.Now()
		err := c.serve(ctx, interval, h.Heartbeat, h.Connected, h.Handle)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !shouldReconnect(err) {
			return err
		}
		c.hook.emit(Event{Direction: "connection", Operation: "RECONNECT", Outcome: "retry"})
		if observer, ok := h.(interface{ Reconnecting() }); ok {
			observer.Reconnecting()
		}
		if time.Since(started) >= time.Minute {
			backoff = time.Second
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		backoff = min(30*time.Second, 2*backoff)
	}
}

type transportError struct{ error }

func (e *transportError) Unwrap() error { return e.error }

func shouldReconnect(err error) bool {
	var app *handlerError
	var transport *transportError
	if errors.As(err, &app) && !errors.As(err, &transport) {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 408 || httpErr.StatusCode == 429 || httpErr.StatusCode >= 500
	}
	var closeErr websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.StatusPolicyViolation:
			return closeErr.Reason == "session initialization busy"
		case 4001, websocket.StatusProtocolError, websocket.StatusUnsupportedData, websocket.StatusInvalidFramePayloadData, websocket.StatusMessageTooBig:
			return false
		}
	}
	return true
}
