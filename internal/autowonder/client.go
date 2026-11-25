package autowonder

import (
	"context"
	"errors"

	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
)

type Config struct {
	BaseURL    string // Platform origin, optionally with a reverse proxy base path.
	ExecutorID int64
	Token      string
	Hostname   string
	Capacity   int // Zero omits the handshake parameter.
	HTTPClient *http.Client
}
type Client struct {
	base            url.URL
	id              int64
	token, hostname string
	capacity        int
	http            http.Client
}

func New(c Config) (*Client, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, errors.New("expected an HTTP(S) platform base URL")
	}
	if c.ExecutorID <= 0 || strings.TrimSpace(c.Token) == "" {
		return nil, errors.New("executor ID and token are required")
	}
	if c.Capacity < 0 || c.Capacity > 50 {
		return nil, errors.New("capacity must be between 1 and 50, or omitted")
	}
	var h http.Client
	if c.HTTPClient != nil {
		h = *c.HTTPClient
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = strings.TrimRight(u.RawPath, "/")
	return &Client{*u, c.ExecutorID, c.Token, c.Hostname, c.Capacity, h}, nil
}

type Session struct {
	conn       *websocket.Conn
	executorID int64
}

// Dial performs the HTTP upgrade. Authentication failure can arrive in a close frame after Dial returns.
func (c *Client) Dial(ctx context.Context) (*Session, error) {
	u := c.base.JoinPath("ws/executor")
	q := u.Query()
	q.Set("executorId", strconv.FormatInt(c.id, 10))
	q.Set("token", c.token)
	if c.hostname != "" {
		q.Set("hostname", c.hostname)
	}
	if c.capacity != 0 {
		q.Set("maxConcurrentDispatches", strconv.Itoa(c.capacity))
	}
	u.RawQuery = q.Encode()
	conn, res, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{HTTPClient: &c.http})
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if res != nil {
			return nil, &HTTPError{StatusCode: res.StatusCode}
		}
		return nil, errors.New("executor connection failed")
	}
	conn.SetReadLimit(16 << 20)
	return &Session{conn: conn, executorID: c.id}, nil
}
func (s *Session) Close() error { return s.conn.CloseNow() }
func (s *Session) Send(ctx context.Context, msg Outgoing) error {
	data, err := Encode(msg)
	if err != nil {
		return &handlerError{err}
	}
	return s.write(ctx, data)
}
func (s *Session) write(ctx context.Context, data []byte) error {
	err := s.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		return &transportError{err}
	}
	return nil
}
func (s *Session) Receive(ctx context.Context) (Incoming, error) {
	kind, data, err := s.conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	if kind != websocket.MessageText {
		return nil, errors.New("expected WebSocket text frame")
	}
	msg, err := decodeForExecutor(data, s.executorID)
	if err != nil {
		return nil, &handlerError{err}
	}
	return msg, nil
}

// Serve runs one authenticated identity until disconnect or cancellation. The owner
// controls reconnect and replay because only it knows which operations are durable.
// Handle should hand long-running work to the executor and return promptly.
func (c *Client) Serve(ctx context.Context, interval time.Duration, heartbeat func() Heartbeat, handle func(context.Context, *Session, Incoming) error) error {
	return c.serve(ctx, interval, heartbeat, nil, handle)
}

func (c *Client) serve(ctx context.Context, interval time.Duration, heartbeat func() Heartbeat, connected func(context.Context, *Session) error, handle func(context.Context, *Session, Incoming) error) error {
	if interval <= 0 || interval >= 90*time.Second || heartbeat == nil || handle == nil {
		return errors.New("expected callbacks and a heartbeat interval below 90s")
	}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	s, err := c.Dial(ctx)
	if err != nil {
		return err
	}
	defer s.Close()
	sendHeartbeat := func() error { h := heartbeat(); return s.Send(ctx, &h) }
	if err := sendHeartbeat(); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := sendHeartbeat(); err != nil {
					cancel(err)
					return
				}
			}
		}
	}()
	defer func() { cancel(nil); <-done }()
	if connected != nil {
		if err := connected(ctx, s); err != nil {
			return &handlerError{err}
		}
	}
	for {
		msg, err := s.Receive(ctx)
		if err != nil {
			if cause := context.Cause(ctx); cause != nil {
				return cause
			}
			return err
		}
		if err = handle(ctx, s, msg); err != nil {
			return &handlerError{err}
		}
	}
}
