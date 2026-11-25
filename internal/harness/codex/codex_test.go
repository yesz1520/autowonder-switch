package codex

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func TestInstalledCodex(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login and model access")
	}
	ctx, stop := context.WithTimeout(t.Context(), 2*time.Minute)
	defer stop()
	b := Backend{Binary: "codex"}
	v, err := b.Version(ctx)
	if err != nil || v == "" {
		t.Fatalf("version: %v", err)
	}
	c, err := b.Models(ctx)
	if err != nil || len(c.Models) == 0 {
		t.Fatalf("catalog: %v", err)
	}
	req := harness.Request{Directory: t.TempDir(), Prompt: "Reply with just the word violet."}
	var turn string
	reply, err := b.Run(ctx, req, harness.Hooks{TurnStarted: func(_, t string) error { turn = t; return nil }})
	if err != nil || reply.Session == "" {
		t.Fatalf("start: %v", err)
	}
	history, status, err := b.ReadTurn(ctx, req.Directory, reply.Session, turn)
	if err != nil || status != "completed" || history.Text != reply.Text {
		t.Fatalf("history status=%q text=%q err=%v", status, history.Text, err)
	}
	req.Session = reply.Session
	req.Prompt = "What word did I ask for? Reply with just that word."
	next, err := b.Run(ctx, req, harness.Hooks{})
	if err != nil || next.Session != reply.Session || !strings.Contains(strings.ToLower(next.Text), "violet") {
		t.Fatalf("resume: %v (%q)", err, next.Text)
	}
	req.Fork = true
	var steered error
	fork, err := b.Run(ctx, req, harness.Hooks{Ready: func(f func(context.Context, string) error) { steered = f(ctx, "Reply with forest.") }})
	if err != nil || steered != nil || fork.Session == reply.Session {
		t.Fatalf("fork: %v steer: %v", err, steered)
	}
}
func TestInstalledCodexWorkspace(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login and model access")
	}
	ctx, stop := context.WithTimeout(t.Context(), 90*time.Second)
	defer stop()
	dir := t.TempDir()
	_, err := (Backend{Binary: "codex"}).Run(ctx, harness.Request{Directory: dir, Writable: true, Environment: map[string]string{"AW_SMOKE_COLOR": "violet"}, Prompt: "Use a shell command to write the value of environment variable AW_SMOKE_COLOR to smoke.txt in the current working directory. Then reply done."}, harness.Hooks{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dir + "/smoke.txt")
	if err != nil || strings.TrimSpace(string(b)) != "violet" {
		t.Fatalf("workspace edit: %v %q", err, b)
	}
}

func TestInstalledCodexPlatformTool(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login and model access")
	}
	ctx, stop := context.WithTimeout(t.Context(), 90*time.Second)
	defer stop()
	var used atomic.Bool
	ready := make(chan func(context.Context, string) error, 1)
	steered := make(chan error, 1)
	spec := harness.Request{Directory: t.TempDir(), Prompt: "Call the autowonder tool with operation list and report its marker value.", Platform: func(toolCtx context.Context, _ string, _ map[string]any) (string, error) {
		if used.CompareAndSwap(false, true) {
			select {
			case steer := <-ready:
				bounded, cancel := context.WithTimeout(toolCtx, 5*time.Second)
				defer cancel()
				steered <- steer(bounded, "Include the marker returned by the current tool in your reply.")
			case <-toolCtx.Done():
				steered <- toolCtx.Err()
			}
		}
		return `{"tools":[],"marker":"swatch246"}`, nil
	}}
	reply, err := (Backend{Binary: "codex"}).Run(ctx, spec, harness.Hooks{Ready: func(steer func(context.Context, string) error) { ready <- steer }})
	if err != nil || !used.Load() || !strings.Contains(reply.Text, "swatch246") {
		t.Fatalf("dynamic tool: %v used=%v reply=%q", err, used.Load(), reply.Text)
	}
	if err := <-steered; err != nil {
		t.Fatalf("steering during dynamic tool: %v", err)
	}
}

func TestInstalledCodexFollowup(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login and model access")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Second)
	defer cancel()
	var written atomic.Bool
	spec := harness.Request{Directory: t.TempDir(), Prompt: "Use request_user_input now to ask me which color I want, with options teal and yellow. Do not guess. After receiving my answer, reply with my chosen color.", Ask: func(_ context.Context, _ string, schema map[string]any) (harness.Answer, error) {
		a := map[string]any{}
		for _, id := range schema["required"].([]string) {
			a[id] = "teal"
		}
		return harness.Answer{Action: "accept", Content: a, Written: func(ok bool) { written.Store(ok) }}, nil
	}}
	reply, err := (Backend{Binary: "codex"}).Run(ctx, spec, harness.Hooks{})
	if err != nil || !written.Load() || !strings.Contains(strings.ToLower(reply.Text), "teal") {
		t.Fatalf("followup %v written=%t result=%q", err, written.Load(), reply.Text)
	}
}

func TestInstalledCodexDeclinedPermission(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_CODEX") != "1" {
		t.Skip("requires local Codex login and model access")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	var written atomic.Bool
	spec := harness.Request{Directory: t.TempDir(), Writable: true, Approvals: true, Prompt: "Use request_permissions to ask for write access to this test directory: " + t.TempDir() + ". Do not touch the directory, do not run commands. After the decision, just say done.", Ask: func(_ context.Context, msg string, _ map[string]any) (harness.Answer, error) {
		if !strings.Contains(msg, "requestApproval") {
			return harness.Answer{}, fmt.Errorf("expected native approval")
		}
		return harness.Answer{Action: "decline", Written: func(ok bool) { written.Store(ok) }}, nil
	}}
	_, err := (Backend{Binary: "codex"}).Run(ctx, spec, harness.Hooks{})
	if err != nil || !written.Load() {
		t.Fatalf("declined permission: %v written=%v", err, written.Load())
	}
}
