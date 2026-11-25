package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func TestConversationRepositories(t *testing.T) {
	fix := newGitFixture(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "capabilities.zip")
	repos := []repository{{Name: "eager", Mode: "eager"}, {Name: "lazy", Mode: "lazy", PinnedBaseCommit: fix.base}}
	raw, err := json.Marshal(repos)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p, archive(t, map[string]string{"repos.json": string(raw)}), 0600); err != nil {
		t.Fatal(err)
	}
	r := runtime{sources: map[string]string{"eager": fix.remote, "lazy": fix.remote}}
	spec := harness.Request{Directory: t.TempDir()}
	cleanup, err := r.openConversationRepos(t.Context(), p, dir, &spec)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	paths, err := filepath.Glob(filepath.Join(dir, "turn-repositories-*", "checkout-*"))
	if err != nil || len(paths) != 1 || !spec.Readable || spec.Writable {
		t.Fatalf("checkouts %v %v", paths, err)
	}
	if _, err = spec.Platform(t.Context(), "executor.open_repository", map[string]any{"name": "other"}); err == nil {
		t.Fatal("unbound repository opened")
	}
	open := func() string {
		v, err := spec.Platform(t.Context(), "executor.open_repository", map[string]any{"name": "lazy"})
		if err != nil {
			t.Fatal(err)
		}
		var out struct{ Path string }
		if err = json.Unmarshal([]byte(v), &out); err != nil {
			t.Fatal(err)
		}
		return out.Path
	}
	if os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
		spec.Prompt = "Open the bound lazy repository via executor.open_repository. Read its README.md using the shell and reply only with the file contents."
		reply, err := (codex.Backend{Binary: "codex"}).Run(t.Context(), spec, harness.Hooks{})
		if err != nil || strings.TrimSpace(reply.Text) != "baseline" {
			t.Fatalf("native repo inspection: %q %v", reply.Text, err)
		}
	}
	path := open()
	if path != open() {
		t.Fatal("lazy repository cloned twice")
	}
	if b, err := os.ReadFile(filepath.Join(path, "README.md")); err != nil || string(b) != "baseline\n" {
		t.Fatalf("repository content %q, %v", b, err)
	}
	cleanup()
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("turn repository retained: %v", err)
	}
}
