// Package mcpclient connects executor-owned MCP sessions for a single dispatch or probe.
package mcpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	Name, Transport, Command, URL                 string
	Args                                          []string
	Env, Headers, EnvSecretRefs, HeaderSecretRefs map[string]string
	Required                                      bool
	TimeoutSeconds                                int64
}

func values(literal, refs, secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(literal)+len(refs))
	for k, v := range literal {
		out[k] = v
	}
	for k, ref := range refs {
		v, ok := secrets[ref]
		if !ok {
			return nil, errors.New("missing MCP secret reference")
		}
		out[k] = v
	}
	return out, nil
}

type originTransport struct {
	origin  *url.URL
	headers map[string]string
}

func (t originTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.EqualFold(req.URL.Scheme, t.origin.Scheme) || !strings.EqualFold(req.URL.Host, t.origin.Host) {
		return nil, errors.New("MCP request changed origin")
	}
	req = req.Clone(req.Context())
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

type Session struct {
	cancel  context.CancelFunc
	timeout time.Duration
	Client  *mcp.ClientSession
	Tools   []*mcp.Tool
}

func (s *Session) Call(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error) {
	call, stop := context.WithTimeout(ctx, s.timeout)
	defer stop()
	return s.Client.CallTool(call, &mcp.CallToolParams{Name: name, Arguments: args})
}

func (s *Session) Close() { defer s.cancel(); s.Client.Close() }

func Open(ctx context.Context, cfg Server, secrets map[string]string, dir, binary string) (*Session, error) {
	var transport mcp.Transport
	switch strings.ToLower(cfg.Transport) {
	case "stdio":
		if cfg.Command == "" {
			return nil, errors.New("missing MCP command")
		}
		env, err := values(cfg.Env, cfg.EnvSecretRefs, secrets)
		if err != nil {
			return nil, err
		}
		if binary == "" {
			binary = "codex"
		}
		args := []string{"sandbox", "-c", `sandbox_mode="workspace-write"`, "--", cfg.Command}
		cmd := exec.CommandContext(ctx, binary, append(args, cfg.Args...)...)
		cmd.Dir = dir
		cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
		for k, v := range env {
			if k == "" || strings.ContainsAny(k, "=\x00") {
				return nil, errors.New("invalid MCP environment key")
			}
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		transport = &mcp.CommandTransport{Command: cmd, TerminateDuration: time.Second}
	case "http", "streamable-http", "streamable_http", "sse":
		u, err := url.Parse(cfg.URL)
		if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return nil, errors.New("invalid MCP HTTP endpoint")
		}
		headers, err := values(cfg.Headers, cfg.HeaderSecretRefs, secrets)
		if err != nil {
			return nil, err
		}
		httpClient := &http.Client{Transport: originTransport{u, headers}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		if strings.EqualFold(cfg.Transport, "sse") {
			transport = &mcp.SSEClientTransport{Endpoint: cfg.URL, HTTPClient: httpClient}
		} else {
			transport = &mcp.StreamableClientTransport{Endpoint: cfg.URL, HTTPClient: httpClient, DisableStandaloneSSE: true}
		}
	default:
		return nil, fmt.Errorf("unsupported MCP transport %q", cfg.Transport)
	}
	seconds := cfg.TimeoutSeconds
	if seconds <= 0 {
		seconds = 60
	}
	// SSE owns its initial GET for the entire session. Limit the handshake
	// without canceling that GET after discovery succeeds.
	probe, cancel := context.WithCancel(ctx)
	timer := time.AfterFunc(time.Duration(min(seconds, 600))*time.Second, cancel)
	defer timer.Stop()
	client := mcp.NewClient(&mcp.Implementation{Name: "autowonder-switch", Version: "1"}, nil)
	s, err := client.Connect(probe, transport, nil)
	if err != nil {
		cancel()
		return nil, errors.New("MCP initialization failed")
	}
	result := &Session{Client: s, cancel: cancel, timeout: time.Duration(min(seconds, 600)) * time.Second}
	for tool, err := range s.Tools(probe, nil) {
		if err != nil {
			result.Close()
			return nil, errors.New("MCP tool discovery failed")
		}
		result.Tools = append(result.Tools, tool)
	}
	if !timer.Stop() || probe.Err() != nil {
		result.Close()
		return nil, errors.New("MCP discovery timed out")
	}
	return result, nil
}
