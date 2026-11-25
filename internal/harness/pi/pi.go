// Package pi runs the local pi coding agent over its JSON-lines RPC interface.
package pi

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/toolbridge"
)

//go:embed tools.ts
var extension []byte

type Backend struct{ Binary string }

func (Backend) Provider() string { return "pi" }
func (b Backend) command(ctx context.Context, args ...string) *exec.Cmd {
	binary := b.Binary
	if binary == "" {
		binary = "pi"
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.WaitDelay = 2 * time.Second
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "AUTOWONDER_TOKEN=") && !strings.HasPrefix(entry, "AUTOWONDER_PI_BRIDGE=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	return cmd
}
func (b Backend) Version(ctx context.Context) (string, error) {
	out, err := b.command(ctx, "--version").Output()
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "", errors.New("empty pi version")
	}
	return v, nil
}

type model struct{ Provider, ID, Name string }

func (m model) qualified() string {
	if m.ID == "" {
		return ""
	}
	if m.Provider == "" {
		return m.ID
	}
	return m.Provider + "/" + m.ID
}

type frame struct {
	Type, ID string
	Success  bool
	Data     json.RawMessage
	Message  struct {
		Role, Model, StopReason, ErrorMessage string
		Content                               json.RawMessage
		Usage                                 struct{ Input, Output, CacheRead, CacheWrite int64 }
	}
}
type process struct {
	cmd  *exec.Cmd
	enc  *json.Encoder
	mu   sync.Mutex
	scan *bufio.Scanner
	stop context.CancelFunc
}

func (b Backend) start(ctx context.Context, dir, bridge string, args ...string) (*process, error) {
	ctx, cancel := context.WithCancel(ctx)
	flags := []string{"--offline", "--mode", "rpc", "--no-extensions", "--no-skills", "--no-prompt-templates", "--no-mcp"}
	cmd := b.command(ctx, append(flags, args...)...)
	cmd.Dir = dir
	if bridge != "" {
		cmd.Env = append(cmd.Env, "AUTOWONDER_PI_BRIDGE="+bridge)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	scan := bufio.NewScanner(out)
	scan.Buffer(make([]byte, 16384), 32<<20)
	return &process{cmd: cmd, enc: json.NewEncoder(stdin), scan: scan, stop: cancel}, nil
}
func (p *process) close() { p.stop(); p.cmd.Wait() }
func (p *process) send(command any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enc.Encode(command)
}
func (p *process) next() (frame, error) {
	var f frame
	if !p.scan.Scan() {
		if err := p.scan.Err(); err != nil {
			return f, err
		}
		return f, errors.New("pi RPC closed")
	}
	err := json.Unmarshal(p.scan.Bytes(), &f)
	return f, err
}
func (p *process) query(kind string, out any) error {
	if err := p.send(map[string]string{"id": kind, "type": kind}); err != nil {
		return err
	}
	for {
		f, err := p.next()
		if err != nil {
			return err
		}
		if f.Type == "response" && f.ID == kind {
			if !f.Success {
				return fmt.Errorf("pi rejected %s", kind)
			}
			return json.Unmarshal(f.Data, out)
		}
	}
}
func (b Backend) Models(ctx context.Context) (harness.Catalog, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	var cat harness.Catalog
	p, err := b.start(ctx, os.TempDir(), "", "--no-tools", "--no-session", "--no-context-files")
	if err != nil {
		return cat, err
	}
	defer p.close()
	var state struct{ Model model }
	if err = p.query("get_state", &state); err != nil {
		return cat, err
	}
	var list struct{ Models []model }
	if err = p.query("get_available_models", &list); err != nil {
		return cat, err
	}
	cat.Default = state.Model.qualified()
	for _, m := range list.Models {
		if id := m.qualified(); id != "" {
			cat.Models = append(cat.Models, harness.Model{ID: id, Name: m.Name})
		}
	}
	return cat, nil
}

var sessionID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func imageInput(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	const limit = 20 << 20
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	typ := http.DetectContentType(data)
	if len(data) > limit || !strings.HasPrefix(typ, "image/") {
		return nil, errors.New("invalid pi image attachment")
	}
	return map[string]string{"type": "image", "data": base64.StdEncoding.EncodeToString(data), "mimeType": typ}, nil
}
func (b Backend) Run(ctx context.Context, q harness.Request, h harness.Hooks) (harness.Reply, error) {
	var reply harness.Reply
	if len(q.ToolHookCommands) > 0 {
		return reply, errors.New("pi package tool hooks unsupported")
	}
	if q.Session != "" && !sessionID.MatchString(q.Session) {
		return reply, errors.New("invalid pi session ID")
	}
	bridge, closeBridge, err := toolbridge.Open(ctx, q)
	if err != nil {
		return reply, err
	}
	defer closeBridge()
	args := []string{}
	if bridge != "" {
		dir, err := os.MkdirTemp("", "autowonder-pi-")
		if err != nil {
			return reply, err
		}
		defer os.RemoveAll(dir)
		file := filepath.Join(dir, "tools.ts")
		if err = os.WriteFile(file, extension, 0600); err != nil {
			return reply, err
		}
		args = append(args, "--tools", "switch_*", "-e", file)
	} else {
		args = append(args, "--no-tools")
	}
	if q.MemoryMode != "provider-local" {
		args = append(args, "--no-context-files")
	}
	if q.SessionDirectory != "" {
		if err = os.MkdirAll(q.SessionDirectory, 0700); err != nil {
			return reply, err
		}
		args = append(args, "--session-dir", q.SessionDirectory)
	}
	if q.Session != "" {
		flag := "--session"
		if q.Fork {
			flag = "--fork"
		}
		args = append(args, flag, q.Session)
	}
	if q.Model != "" {
		args = append(args, "--model", q.Model)
	}
	instructions := q.Instructions
	if q.OutputSchema != nil {
		schema, err := json.Marshal(q.OutputSchema)
		if err != nil {
			return reply, err
		}
		instructions += "\nReturn only JSON matching this schema: " + string(schema)
	}
	if instructions != "" {
		args = append(args, "--append-system-prompt", instructions)
	}
	p, err := b.start(ctx, q.Directory, bridge, args...)
	if err != nil {
		return reply, err
	}
	defer p.close()
	var state struct {
		SessionID string
		Model     model
	}
	if err = p.query("get_state", &state); err != nil {
		return reply, err
	}
	reply.Session = state.SessionID
	reply.Model = state.Model.qualified()
	if !sessionID.MatchString(reply.Session) || (q.Session != "" && (reply.Session == q.Session) == q.Fork) {
		return reply, errors.New("pi returned unexpected session")
	}
	if q.Model != "" && q.Model != reply.Model && q.Model != state.Model.ID {
		return reply, errors.New("pi did not select requested model")
	}
	if h.Session != nil {
		if err = h.Session(reply.Session); err != nil {
			return reply, err
		}
	}
	prompt := map[string]any{"id": "prompt", "type": "prompt", "message": q.Prompt}
	if len(q.Images) > 0 {
		var imgs []map[string]string
		for _, path := range q.Images {
			v, e := imageInput(path)
			if e != nil {
				return reply, e
			}
			imgs = append(imgs, v)
		}
		prompt["images"] = imgs
	}
	if err = p.send(prompt); err != nil {
		return reply, err
	}
	// Steering acknowledgements share stdout with task events. Correlate them by
	// request ID so a late acknowledgement cannot resolve a later request.
	var pending sync.Map
	var sequence atomic.Uint64
	done := make(chan struct{})
	defer close(done)
	steer := func(call context.Context, text string) error {
		id := fmt.Sprintf("steer-%d", sequence.Add(1))
		ack := make(chan bool, 1)
		pending.Store(id, ack)
		defer pending.Delete(id)
		select {
		case <-done:
			return errors.New("pi turn finished")
		case <-call.Done():
			return call.Err()
		default:
		}
		if err := p.send(map[string]string{"type": "steer", "id": id, "message": text}); err != nil {
			return err
		}
		select {
		case ok := <-ack:
			if !ok {
				return errors.New("pi rejected steering")
			}
			return nil
		case <-done:
			return errors.New("pi turn finished")
		case <-call.Done():
			return call.Err()
		}
	}
	reply.Usage = &harness.Usage{Provider: "pi", Model: reply.Model}
	for {
		f, err := p.next()
		if ctx.Err() != nil {
			return reply, ctx.Err()
		}
		if err != nil {
			return reply, err
		}
		switch f.Type {
		case "agent_start":
			if h.Ready != nil {
				h.Ready(steer)
			}
		case "response":
			if ack, ok := pending.Load(f.ID); ok {
				ack.(chan bool) <- f.Success
				continue
			}
			if strings.HasPrefix(f.ID, "steer-") {
				continue
			}
			if !f.Success {
				return reply, errors.New("pi rejected RPC command")
			}
		case "message_end":
			m := f.Message
			if m.Role != "assistant" {
				continue
			}
			if m.StopReason == "error" || m.StopReason == "aborted" {
				if strings.EqualFold(strings.TrimSuffix(m.ErrorMessage, "."), "connection error") {
					return reply, errors.New("pi could not connect to the configured model endpoint")
				}
				return reply, fmt.Errorf("pi model turn ended: %s", m.StopReason)
			}
			var text strings.Builder
			var parts []struct{ Type, Text string }
			if err := json.Unmarshal(m.Content, &parts); err != nil {
				return reply, err
			}
			for _, part := range parts {
				if part.Type == "text" {
					text.WriteString(part.Text)
				}
			}
			reply.Text = text.String()
			if reply.Text != "" && h.Message != nil {
				h.Message(reply.Text)
			}
			u := reply.Usage
			u.Input += m.Usage.Input
			u.Output += m.Usage.Output
			u.CacheRead += m.Usage.CacheRead
			u.CacheWrite += m.Usage.CacheWrite
		case "agent_end":
			if strings.TrimSpace(reply.Text) == "" {
				return reply, errors.New("pi returned empty reply")
			}
			return reply, nil
		}
	}
}
