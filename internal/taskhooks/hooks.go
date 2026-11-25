// Package taskhooks loads verified AutoWonder hook bundles for one dispatch.
package taskhooks

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type descriptor struct {
	Schema      string `yaml:"schemaVersion"`
	Name        string
	Trigger     string
	Interpreter string
	Command     string
	Cwd         string
	Timeout     int    `yaml:"timeoutSeconds"`
	Failure     string `yaml:"failurePolicy"`
	Once        string `yaml:"onceScope"`
}

type Set struct {
	mu              sync.Mutex
	base, workspace string
	binary          string
	hooks           []descriptor
}

var component = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

func Load(pkg, dir, workspace string, attempt int64, binary string) (*Set, error) {
	z, err := zip.OpenReader(pkg)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	f, err := z.Open("hooks.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var index struct {
		Schema string `json:"schemaVersion"`
		Hooks  []struct {
			Name, SHA256 string
			Enabled      *bool
		}
	}
	if err = json.NewDecoder(f).Decode(&index); err != nil {
		return nil, err
	}
	if index.Schema != "autowonder.hooks.v1" {
		return nil, errors.New("unsupported hook index")
	}
	if binary == "" {
		binary = "codex"
	}
	set := &Set{binary: binary, base: filepath.Join(dir, fmt.Sprintf("hooks-%d", attempt)), workspace: workspace}
	seen := map[string]bool{}
	for _, v := range index.Hooks {
		if v.Enabled != nil && !*v.Enabled {
			continue
		}
		if !component.MatchString(v.Name) || seen[v.Name] {
			return nil, errors.New("invalid or repeated hook name")
		}
		seen[v.Name] = true
		h, err := extract(&z.Reader, set.base, v.Name, v.SHA256)
		if err != nil {
			return nil, fmt.Errorf("hook %s: %w", v.Name, err)
		}
		set.hooks = append(set.hooks, h)
	}
	return set, nil
}

func extract(z *zip.Reader, base, name, digest string) (descriptor, error) {
	var h descriptor
	if err := os.MkdirAll(base, 0700); err != nil {
		return h, err
	}
	temp, err := os.MkdirTemp(base, ".hook-")
	if err != nil {
		return h, err
	}
	defer os.RemoveAll(temp)
	prefix := "capabilities/hooks/" + name + "/"
	var lines []string
	var total uint64
	for _, f := range z.File {
		path, ok := strings.CutPrefix(f.Name, prefix)
		if !ok || f.FileInfo().IsDir() {
			continue
		}
		total += f.UncompressedSize64
		if !f.Mode().IsRegular() || !fs.ValidPath(path) || strings.Contains(path, "\\") || total > 50<<20 || len(lines) >= 500 {
			return h, errors.New("invalid hook package entry")
		}
		b, err := fs.ReadFile(z, f.Name)
		if err != nil {
			return h, err
		}
		if path == "hook.yaml" {
			if err = yaml.Unmarshal(b, &h); err != nil {
				return h, err
			}
		}
		dest := filepath.Join(temp, path)
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return h, err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return h, err
		}
		_, werr := out.Write(b)
		if err = errors.Join(werr, out.Close()); err != nil {
			return h, err
		}
		lines = append(lines, fmt.Sprintf("%s\x00sha256:%x\n", path, sha256.Sum256(b)))
	}
	sort.Strings(lines)
	if fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(strings.Join(lines, "")))) != digest {
		return h, errors.New("hook bundle checksum mismatch")
	}
	if h.Schema != "autowonder.hook.v1" || h.Name != name || !fs.ValidPath(h.Command) || strings.Contains(h.Command, "\\") {
		return h, errors.New("invalid hook descriptor")
	}
	if _, err = os.Stat(filepath.Join(temp, h.Command)); err != nil {
		return h, err
	}
	if !slices.Contains([]string{"sh", "bash", "python3", "node"}, h.Interpreter) {
		return h, fmt.Errorf("unsupported hook interpreter %q", h.Interpreter)
	}
	if h.Timeout == 0 {
		h.Timeout = 30
	}
	if h.Once == "" {
		h.Once = "none"
	}
	if h.Cwd == "" {
		h.Cwd = "workspace"
	}
	if h.Failure == "" {
		h.Failure = "fail-task"
	}
	if h.Timeout < 1 || h.Timeout > 86400 || !slices.Contains([]string{"none", "once-per-attempt"}, h.Once) || !slices.Contains([]string{"fail-task", "continue"}, h.Failure) || !slices.Contains([]string{"workspace", "triggeringRepo"}, h.Cwd) {
		return h, errors.New("unsupported hook execution policy")
	}
	if !slices.Contains([]string{"beforeTool", "afterTool", "beforeRepoPrepare", "afterRepoPrepare", "beforeAgentStart", "afterAgentExit", "beforeStep", "afterStep", "beforeCommit", "beforePush", "onFailure", "cleanup"}, h.Trigger) {
		return h, fmt.Errorf("unsupported hook trigger %q", h.Trigger)
	}
	// Restore the verified bytes on each load; do not reuse an unchecked copy.
	bundle := filepath.Join(base, "bundle-"+name)
	if err = os.RemoveAll(bundle); err != nil {
		return h, err
	}
	return h, os.Rename(temp, bundle)
}

func (s *Set) Has(event string) bool {
	if s == nil {
		return false
	}
	for _, h := range s.hooks {
		if h.Trigger == event {
			return true
		}
	}
	return false
}
func (s *Set) Run(ctx context.Context, event, repo string, fields map[string]any) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, h := range s.hooks {
		if h.Trigger != event {
			continue
		}
		marker := filepath.Join(s.base, "done-"+h.Name)
		if h.Once == "once-per-attempt" {
			if _, err := os.Stat(marker); err == nil {
				continue
			} else if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
		cwd := s.workspace
		if h.Cwd == "triggeringRepo" {
			if repo == "" {
				return errors.New("hook requires a triggering repository")
			}
			cwd = repo
		}
		input := map[string]any{"trigger": event, "provider": "codex", "cwd": cwd}
		for k, v := range fields {
			input[k] = v
		}
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		logDir := filepath.Join(s.base, "logs", event, h.Name)
		if err = os.MkdirAll(logDir, 0700); err != nil {
			return err
		}
		if err = runScript(ctx, s.binary, h, filepath.Join(s.base, "bundle-"+h.Name, h.Command), cwd, logDir, b); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if h.Failure == "continue" {
				continue
			}
			return fmt.Errorf("%s hook %s: %w", event, h.Name, err)
		}
		if h.Once == "once-per-attempt" {
			if err = os.WriteFile(marker, nil, 0600); err != nil {
				return err
			}
		}
	}
	return nil
}

func runScript(ctx context.Context, binary string, h descriptor, path, cwd, logs string, input []byte) error {
	ctx, stop := context.WithTimeout(ctx, time.Duration(h.Timeout)*time.Second)
	defer stop()
	cmd := exec.CommandContext(ctx, binary, "sandbox", "-c", `sandbox_mode="workspace-write"`, "-c", "sandbox_workspace_write.network_access=false", "--", h.Interpreter, path)
	cmd.Dir = cwd
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "LANG=" + os.Getenv("LANG")}
	cmd.Stdin = bytes.NewReader(input)
	stdout, err := os.OpenFile(filepath.Join(logs, "stdout"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer stdout.Close()
	stderr, err := os.OpenFile(filepath.Join(logs, "stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer stderr.Close()
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err = cmd.Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
