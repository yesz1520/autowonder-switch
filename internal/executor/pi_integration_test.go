package executor

import (
	"context"
	"encoding/json"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/pi"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func piFixture(t *testing.T) {
	t.Helper()
	claudeFixture(t)
	dir := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	provider := map[string]any{"api": "anthropic-messages", "baseUrl": os.Getenv("ANTHROPIC_BASE_URL"), "apiKey": "fixture", "models": []map[string]string{{"id": "fixture"}}}
	models, err := json.Marshal(map[string]any{"providers": map[string]any{"localfixture": provider}})
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"models.json": models, "settings.json": []byte(`{"defaultProvider":"localfixture","defaultModel":"fixture"}`)} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPiNativeWorkspaceAndFork(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_PI") != "1" {
		t.Skip("requires installed pi CLI")
	}
	piFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()
	dir := t.TempDir()
	q := harness.Request{Directory: dir, SessionDirectory: t.TempDir(), Writable: true, Prompt: "first-turn-marker workspace-test"}
	backend := pi.Backend{}
	first, err := backend.Run(ctx, q, harness.Hooks{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "artifact.txt"))
	if err != nil || string(b) != "verified\n" {
		t.Fatalf("missing workspace artifact: %v", err)
	}
	q.Directory = t.TempDir()
	q.Fork = true
	q.Session = first.Session
	q.Prompt = "confirm-prior-turn"
	fork, err := backend.Run(ctx, q, harness.Hooks{})
	if err != nil || fork.Session == first.Session || fork.Text != "preserved-history" {
		t.Fatalf("fork failed: %+v %v", fork, err)
	}
}

func TestPiNativeSteering(t *testing.T) {
	if os.Getenv("AUTOWONDER_TEST_PI") != "1" {
		t.Skip("requires installed pi CLI")
	}
	piFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	ack := make(chan error, 1)
	var once sync.Once
	hooks := harness.Hooks{Ready: func(steer func(context.Context, string) error) {
		once.Do(func() { go func() { ack <- steer(ctx, "confirm-prior-turn") }() })
	}}
	reply, err := (pi.Backend{}).Run(ctx, harness.Request{Directory: t.TempDir(), SessionDirectory: t.TempDir(), Prompt: "first-turn-marker"}, hooks)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-ack:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if reply.Text != "preserved-history" {
		t.Fatalf("steering did not reach model: %q", reply.Text)
	}
}
