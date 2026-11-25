package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/harness/codex"
)

func TestCodexPackagePlugin(t *testing.T) {
	dir := t.TempDir()
	native := os.Getenv("AUTOWONDER_TEST_CODEX") == "1"
	if !native {
		t.Setenv("CODEX_HOME", t.TempDir())
	}
	pkg := filepath.Join(dir, "p.zip")
	files := map[string]string{
		"workitem.md": "Use the aw-inspection skill to answer with its verification word only.",
		"skills.json": `{"plugins":[{"name":"wrong-provider","path":"capabilities/plugins/wrong-provider","providers":["qoder"],"required":true},{"name":"aw-test","path":"capabilities/plugins/aw-test","providers":["codex"],"required":true}]}`,
		"capabilities/plugins/aw-test/.codex-plugin/plugin.json":     `{"name":"aw-test","version":"1.0.0","skills":"./skills/"}`,
		"capabilities/plugins/aw-test/skills/aw-inspection/SKILL.md": "---\nname: aw-inspection\ndescription: Provides the verification word requested in an inspection.\n---\nThe verification word is mercury9883. Respond with that word only.",
		"capabilities/plugins/wrong-provider/image.bin":              "\xff\x00",
	}
	if err := os.WriteFile(pkg, archive(t, files), 0600); err != nil {
		t.Fatal(err)
	}
	spec := harness.Request{Directory: t.TempDir()}
	var err error
	spec.Prompt, err = packageText(pkg, true, &spec)
	if err != nil || strings.Contains(spec.Prompt, "mercury") {
		t.Fatal("plugin source leaked into prompt", err)
	}
	r := runtime{}
	if _, e := r.openPackageCapabilities(pkg, &spec); e == nil {
		t.Fatal("loaded executable plugin without opt-in")
	}
	r.files.NativePlugins = true
	close, err := r.openPackageCapabilities(pkg, &spec)
	if err != nil {
		t.Fatal(err)
	}
	defer close()
	if len(spec.PluginIDs) != 1 || !strings.HasPrefix(spec.PluginIDs[0], "aw-test@autowonder-") {
		t.Fatal(spec.PluginIDs)
	}
	if native {
		reply, err := (codex.Backend{Binary: "codex"}).Run(t.Context(), spec, harness.Hooks{})
		if err != nil || strings.TrimSpace(reply.Text) != "mercury9883" {
			t.Fatalf("native plugin reply: %q %v", reply.Text, err)
		}
	} else {
		_, market, _ := strings.Cut(spec.PluginIDs[0], "@")
		cache := filepath.Join(os.Getenv("CODEX_HOME"), "plugins/cache", market)
		if _, err = os.Stat(filepath.Join(cache, "aw-test/local/.codex-plugin/plugin.json")); err != nil {
			t.Fatal(err)
		}
		close()
		if _, err = os.Stat(cache); !os.IsNotExist(err) {
			t.Fatal("plugin cache retained", err)
		}
	}
}

func TestPackagedSkillResources(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "capabilities.zip")
	files := map[string]string{
		"skills.json": `{"skills":[{"name":"inspect-swatch","path":"capabilities/skills/inspect-swatch","required":true}]}`,
		"capabilities/skills/inspect-swatch/SKILL.md":   "---\nname: inspect-swatch\ndescription: Inspect the provided swatch.\n---\nRead swatch.bin in this skill's directory. Report its first three bytes as lowercase hex without separators.",
		"capabilities/skills/inspect-swatch/swatch.bin": "\x80\x00\xed",
	}
	if err := os.WriteFile(pkg, archive(t, files), 0600); err != nil {
		t.Fatal(err)
	}
	spec := harness.Request{Directory: t.TempDir(), Prompt: "Use inspect-swatch. Respond only with the requested hex."}
	if _, err := packageText(pkg, false, &spec); err != nil {
		t.Fatal(err)
	}
	r := runtime{}
	cleanup, err := r.openPackageCapabilities(pkg, &spec)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	paths, err := filepath.Glob(filepath.Join(dir, "turn-skills-*", "inspect-swatch", "swatch.bin"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("resources: %v %v", paths, err)
	}
	if b, err := os.ReadFile(paths[0]); err != nil || string(b) != "\x80\x00\xed" {
		t.Fatalf("skill resource: %x %v", b, err)
	}
	if !spec.Readable || !strings.Contains(spec.Instructions, filepath.Join(filepath.Dir(paths[0]), "SKILL.md")) {
		t.Fatal(spec)
	}
	if os.Getenv("AUTOWONDER_TEST_CODEX") == "1" {
		reply, err := (codex.Backend{Binary: "codex"}).Run(t.Context(), spec, harness.Hooks{})
		if err != nil || strings.TrimSpace(reply.Text) != "8000ed" {
			t.Fatalf("native skill reply %q: %v", reply.Text, err)
		}
	}
	cleanup()
	if _, err = os.Stat(paths[0]); !os.IsNotExist(err) {
		t.Fatal("skill files retained", err)
	}
}
