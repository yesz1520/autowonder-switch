// Package claude adapts the installed Claude Code CLI to the executor.
package claude

import (
	"bufio"
	"bytes"
	"context"
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
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/toolbridge"
)

type Backend struct{ Binary string }

func (Backend) Provider() string { return "claude" }
func (b Backend) command(ctx context.Context, args ...string) *exec.Cmd {
	binary := b.Binary
	if binary == "" {
		binary = "claude"
		if _, err := exec.LookPath(binary); err != nil {
			if home, e := os.UserHomeDir(); e == nil {
				binary = filepath.Join(home, ".local", "bin", "claude")
			}
		}
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.WaitDelay = 2 * time.Second
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "AUTOWONDER_TOKEN=") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	return cmd
}
func (b Backend) Version(ctx context.Context) (string, error) {
	data, err := b.command(ctx, "--version").Output()
	if err != nil {
		return "", err
	}
	f := strings.Fields(string(data))
	if len(f) == 0 {
		return "", errors.New("Claude Code returned an empty version")
	}
	return f[0], nil
}

var streamArgs = []string{"-p", "--verbose", "--input-format", "stream-json", "--output-format", "stream-json", "--strict-mcp-config", "--setting-sources", "user"}

func (b Backend) Models(ctx context.Context) (harness.Catalog, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := b.command(ctx, append(append([]string{}, streamArgs...), "--tools", "")...)
	cmd.Dir = os.TempDir()
	in, err := cmd.StdinPipe()
	if err != nil {
		return harness.Catalog{}, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return harness.Catalog{}, err
	}
	if err = cmd.Start(); err != nil {
		return harness.Catalog{}, err
	}
	defer func() { cancel(); in.Close(); cmd.Wait() }()
	if _, err = io.WriteString(in, `{"type":"control_request","request_id":"catalog","request":{"subtype":"initialize"}}
{"type":"control_request","request_id":"selection","request":{"subtype":"get_settings"}}
`); err != nil {
		return harness.Catalog{}, err
	}
	var cat harness.Catalog
	var selected bool
	scan := bufio.NewScanner(out)
	scan.Buffer(make([]byte, 4096), 4<<20)
	for scan.Scan() {
		var v struct {
			Response struct {
				ID       string `json:"request_id"`
				Response struct {
					Models  []struct{ Value, DisplayName string }
					Applied struct{ Model string }
				}
			}
		}
		if json.Unmarshal(scan.Bytes(), &v) != nil {
			continue
		}
		if v.Response.ID == "selection" {
			selected = true
			cat.Default = v.Response.Response.Applied.Model
		}
		for _, m := range v.Response.Response.Models {
			if m.Value != "" {
				cat.Models = append(cat.Models, harness.Model{ID: m.Value, Name: m.DisplayName})
			}
		}
		if len(cat.Models) > 0 && selected {
			if cat.Default == "" {
				cat.Default = cat.Models[0].ID
			}
			return cat, nil
		}
	}
	return harness.Catalog{}, errors.New("Claude Code returned no model catalog")
}

func userMessage(q harness.Request) ([]byte, error) {
	content := []map[string]any{{"type": "text", "text": q.Prompt}}
	for _, path := range q.Images {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(f, (20<<20)+1))
		f.Close()
		if err != nil {
			return nil, err
		}
		if len(b) > 20<<20 {
			return nil, errors.New("Claude image exceeds 20 MiB")
		}
		typ := http.DetectContentType(b)
		switch typ {
		case "image/png", "image/jpeg", "image/webp", "image/gif":
		default:
			return nil, errors.New("unsupported Claude image")
		}
		content = append(content, map[string]any{"type": "image", "source": map[string]string{"type": "base64", "media_type": typ, "data": base64.StdEncoding.EncodeToString(b)}})
	}
	return json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}})
}

var shellVariable = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func shellLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

var apiStatus = regexp.MustCompile(`(?i)API Error:\s*([45][0-9]{2})\b`)

func (b Backend) Run(ctx context.Context, q harness.Request, h harness.Hooks) (harness.Reply, error) {
	var reply harness.Reply
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	args := append([]string{}, streamArgs...)
	if q.Session != "" {
		args = append(args, "--resume", q.Session)
		if q.Fork {
			args = append(args, "--fork-session")
		}
	}
	if q.Model != "" {
		args = append(args, "--model", q.Model)
	}
	if q.Instructions != "" {
		args = append(args, "--append-system-prompt", q.Instructions)
	}
	nativeTools := ""
	if q.Readable || q.Writable {
		nativeTools = "Read,Glob,Grep"
	}
	settings := map[string]any{}
	hooks := map[string]any{}
	if q.Writable {
		nativeTools += ",Bash"
		settings["sandbox"] = map[string]any{"enabled": true, "failIfUnavailable": true, "allowUnsandboxedCommands": false}
		if len(q.Environment) > 0 {
			dir, err := os.MkdirTemp("", "autowonder-claude-env-")
			if err != nil {
				return reply, err
			}
			defer os.RemoveAll(dir)
			var exports strings.Builder
			for name, value := range q.Environment {
				if !shellVariable.MatchString(name) || strings.ContainsRune(value, 0) {
					return reply, errors.New("invalid Claude workspace environment variable")
				}
				fmt.Fprintf(&exports, "export %s=%s\n", name, shellLiteral(value))
			}
			path := filepath.Join(dir, "workspace.env")
			if err := os.WriteFile(path, []byte(exports.String()), 0600); err != nil {
				return reply, err
			}
			hooks["SessionStart"] = []any{map[string]any{"hooks": []any{map[string]string{"type": "command", "command": "cat " + shellLiteral(path) + ` >> "$CLAUDE_ENV_FILE"`}}}}
		}
	}
	args = append(args, "--tools", nativeTools)
	for _, event := range []string{"PreToolUse", "PostToolUse"} {
		if command := q.ToolHookCommands[event]; command != "" {
			hooks[event] = []any{map[string]any{"matcher": "*", "hooks": []any{map[string]any{"type": "command", "command": command, "timeout": 86400}}}}
		}
	}
	if len(hooks) > 0 {
		settings["hooks"] = hooks
	}
	if len(settings) > 0 {
		data, err := json.Marshal(settings)
		if err != nil {
			return reply, err
		}
		args = append(args, "--settings", string(data))
	}
	if q.OutputSchema != nil {
		schema, err := json.Marshal(q.OutputSchema)
		if err != nil {
			return reply, err
		}
		args = append(args, "--json-schema", string(schema))
	}
	bridge, closeBridge, err := toolbridge.Open(ctx, harness.Request{Platform: q.Platform, Ask: q.Ask})
	if err != nil {
		return reply, err
	}
	defer closeBridge()
	if bridge != "" {
		cfg, err := json.Marshal(map[string]any{"mcpServers": map[string]any{"switch": map[string]string{"type": "http", "url": bridge}}})
		if err != nil {
			return reply, err
		}
		args = append(args, "--mcp-config", string(cfg), "--allowedTools", "mcp__switch__*")
	}
	cmd := b.command(ctx, args...)
	cmd.Dir = q.Directory
	if q.MemoryMode != "provider-local" {
		cmd.Env = append(cmd.Env, "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1")
	}
	input, err := userMessage(q)
	if err != nil {
		return reply, err
	}
	cmd.Stdin = bytes.NewReader(append(input, '\n'))
	out, err := cmd.StdoutPipe()
	if err != nil {
		return reply, err
	}
	if err = cmd.Start(); err != nil {
		return reply, err
	}
	defer func() {
		if cmd.ProcessState == nil {
			cancel()
			cmd.Wait()
		}
	}()
	scan := bufio.NewScanner(out)
	scan.Buffer(make([]byte, 16384), 32<<20)
	gotResult := false
	for scan.Scan() {
		var v struct {
			Type, Subtype, Result, Model string
			Session                      string          `json:"session_id"`
			IsError                      bool            `json:"is_error"`
			Structured                   json.RawMessage `json:"structured_output"`
			Message                      struct {
				Model   string
				Content []struct{ Type, Text string }
			}
			Usage struct {
				Input  int64 `json:"input_tokens"`
				Output int64 `json:"output_tokens"`
				Read   int64 `json:"cache_read_input_tokens"`
				Write  int64 `json:"cache_creation_input_tokens"`
			}
		}
		if err = json.Unmarshal(scan.Bytes(), &v); err != nil {
			return reply, err
		}
		switch v.Type {
		case "system":
			if v.Subtype == "init" {
				reply.Session = v.Session
				reply.Model = v.Model
				if q.Session != "" && ((!q.Fork && v.Session != q.Session) || (q.Fork && v.Session == q.Session)) {
					return reply, errors.New("Claude Code returned an unexpected session")
				}
				if h.Session != nil {
					if err = h.Session(v.Session); err != nil {
						return reply, err
					}
				}
			}
		case "assistant":
			if v.Message.Model != "" {
				reply.Model = v.Message.Model
			}
			if h.Message != nil {
				for _, c := range v.Message.Content {
					if c.Type == "text" && c.Text != "" {
						h.Message(c.Text)
					}
				}
			}
		case "result":
			if v.IsError || v.Subtype != "success" {
				if status := apiStatus.FindStringSubmatch(v.Result); len(status) > 1 {
					return reply, fmt.Errorf("Claude Code model API returned HTTP %s", status[1])
				}
				return reply, errors.New("Claude Code did not complete the turn")
			}
			reply.Text = v.Result
			if q.OutputSchema != nil && len(v.Structured) > 0 {
				reply.Text = string(v.Structured)
			}
			reply.Usage = &harness.Usage{Provider: "claude", Model: reply.Model, Input: v.Usage.Input, Output: v.Usage.Output, CacheRead: v.Usage.Read, CacheWrite: v.Usage.Write}
			gotResult = true
		}
	}
	if ctx.Err() != nil {
		return reply, ctx.Err()
	}
	if err = scan.Err(); err != nil {
		return reply, err
	}
	if err = cmd.Wait(); err != nil {
		return reply, err
	}
	if !gotResult || reply.Session == "" {
		return reply, errors.New("incomplete Claude Code response")
	}
	return reply, nil
}
