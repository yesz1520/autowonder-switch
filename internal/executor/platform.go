package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

// The platform authorizes tools with the dispatch/turn credential. The model
// only sees tool schemas, arguments and results; it never receives the token.
func (r *runtime) platform(token string) harness.PlatformTool {
	if token == "" {
		return nil
	}
	return func(ctx context.Context, name string, args map[string]any) (string, error) {
		method := "tools/list"
		var params any = map[string]any{}
		if name != "" {
			method = "tools/call"
			params = map[string]any{"name": name, "arguments": args}
		}
		endpoint, err := url.JoinPath(r.base, "api/mcp")
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
		if err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		c := http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := c.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("platform MCP returned HTTP %d", resp.StatusCode)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20+1))
		if err != nil {
			return "", err
		}
		if len(data) > 4<<20 {
			return "", errors.New("platform tool response too large")
		}
		var reply struct {
			JSONRPC string
			ID      int
			Result  json.RawMessage
			Error   *struct {
				Code    int
				Message string
			}
		}
		if err = json.Unmarshal(data, &reply); err != nil {
			return "", err
		}
		if reply.JSONRPC != "2.0" || reply.ID != 1 {
			return "", errors.New("invalid platform tool response")
		}
		if reply.Error != nil {
			return "", fmt.Errorf("platform tool %d: %s", reply.Error.Code, reply.Error.Message)
		}
		if len(reply.Result) == 0 {
			return "", errors.New("missing platform tool result")
		}
		if name != "" {
			var outcome struct{ IsError bool }
			if err = json.Unmarshal(reply.Result, &outcome); err != nil {
				return "", err
			}
			if outcome.IsError {
				return "", fmt.Errorf("platform tool rejected the call: %s", reply.Result)
			}
		}
		return string(reply.Result), nil
	}
}
