package taskhooks

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T, event, policy, script string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	work := filepath.Join(dir, "workspace")
	if err := os.Mkdir(work, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "package.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	var lines []string
	for name, b := range map[string]string{"hook.yaml": "schemaVersion: autowonder.hook.v1\nname: check\ntrigger: " + event + "\ninterpreter: sh\ncommand: scripts/check.sh\ncwd: workspace\nfailurePolicy: " + policy + "\nonceScope: once-per-attempt\n", "scripts/check.sh": script} {
		w, err := z.Create("capabilities/hooks/check/" + name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprint(w, b)
		lines = append(lines, fmt.Sprintf("%s\x00sha256:%x\n", name, sha256.Sum256([]byte(b))))
	}
	sort.Strings(lines)
	w, err := z.Create("hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": "autowonder.hooks.v1", "hooks": []any{map[string]any{"name": "check", "sha256": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(strings.Join(lines, "")))), "enabled": true}}}); err != nil {
		t.Fatal(err)
	}
	if err = errors.Join(z.Close(), f.Close()); err != nil {
		t.Fatal(err)
	}
	return path, dir, work
}
func TestAttemptHook(t *testing.T) {
	pkg, dir, work := fixture(t, "beforeAgentStart", "fail-task", "echo called >> proof")
	for range 2 {
		s, err := Load(pkg, dir, work, 1, "codex")
		if err != nil {
			t.Fatal(err)
		}
		if err = s.Run(t.Context(), "beforeAgentStart", "", nil); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(work, "proof"))
	if err != nil || string(data) != "called\n" {
		t.Fatalf("proof %q: %v", data, err)
	}
}
func TestInstalledCodexBlocksTool(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires installed Codex and model login")
	}
	ctx, stop := context.WithTimeout(t.Context(), 110*time.Second)
	defer stop()
	pkg, dir, work := fixture(t, "beforeTool", "fail-task", "cat; exit 1")
	s, err := Load(pkg, dir, work, 1, "codex")
	if err != nil {
		t.Fatal(err)
	}
	commands, closeHooks, err := s.ToolCommands(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer closeHooks()
	_, err = (codex.Backend{Binary: "codex"}).Run(ctx, harness.Request{Directory: work, Writable: true, ToolHookCommands: commands, Prompt: "Attempt exactly one shell tool call to run: touch must-not-exist. If denied, stop and answer blocked. Do not try another tool."}, harness.Hooks{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(s.base, "logs", "beforeTool", "check", "stdout"))
	if err != nil || !strings.Contains(string(b), "must-not-exist") || !strings.Contains(string(b), "autowonder.toolHookContext.v1") {
		t.Fatalf("hook stdin %q: %v", b, err)
	}
	if _, err = os.Stat(filepath.Join(work, "must-not-exist")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("tool was not blocked: %v", err)
	}
}

func TestInstalledCodexPostTool(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires installed Codex and model login")
	}
	ctx, stop := context.WithTimeout(t.Context(), 110*time.Second)
	defer stop()
	pkg, dir, work := fixture(t, "afterTool", "continue", "cat; exit 1")
	s, err := Load(pkg, dir, work, 1, "codex")
	if err != nil {
		t.Fatal(err)
	}
	commands, closeHooks, err := s.ToolCommands(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer closeHooks()
	_, err = (codex.Backend{Binary: "codex"}).Run(ctx, harness.Request{Directory: work, Writable: true, ToolHookCommands: commands, Prompt: "Run exactly one shell command: echo ran > proof && cat proof. Reply with its output."}, harness.Hooks{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(s.base, "logs", "afterTool", "check", "stdout"))
	var input struct {
		ToolOutput json.RawMessage
		ToolName   string
	}
	if err != nil || json.Unmarshal(b, &input) != nil || input.ToolName == "" || !strings.Contains(string(input.ToolOutput), "ran") {
		t.Fatalf("hook stdin %q: %v", b, err)
	}
	data, err := os.ReadFile(filepath.Join(work, "proof"))
	if err != nil || string(data) != "ran\n" {
		t.Fatalf("file %q: %v", data, err)
	}
}
