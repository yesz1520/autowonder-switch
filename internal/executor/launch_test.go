package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

func TestLaunchMemoryModes(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "p.zip")
	if err := os.WriteFile(pkg, archive(t, map[string]string{"workitem.md": "read", "memory/topic.md": "platform note amber"}), 0600); err != nil {
		t.Fatal(err)
	}
	client, err := aw.New(aw.Config{BaseURL: "http://localhost", ExecutorID: 1, Token: "executor"})
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"platform", "provider-local", "none"} {
		r := runtime{client: client, launch: launchConfig{MemoryMode: mode}}
		spec := harness.Request{MemoryMode: mode}
		text, err := packageText(pkg, true, &spec)
		if err != nil {
			t.Fatal(err)
		}
		if err = r.bindTaskMemory(&aw.TaskDispatch{DispatchID: 1, MemorySnapshotPath: "/snapshot", DispatchMCPToken: "dispatch"}, &spec); err != nil {
			t.Fatal(err)
		}
		want := mode == "platform"
		if strings.Contains(text, "platform note amber") != want || (spec.Platform != nil) != want {
			t.Fatalf("%s: unexpected memory injection", mode)
		}
	}
}

func TestLaunchVersionPrecedence(t *testing.T) {
	dir := t.TempDir()
	r := runtime{dir: dir, launch: launchConfig{AppliedRequest: "launch-first", RuntimeVersion: "1.3.7"}}
	if err := r.loadUpgrade(); err != nil || r.version != "1.3.7" {
		t.Fatal(r.version, err)
	}
	r.upgrades.Version = "1.3.8"
	if err := save(filepath.Join(dir, "upgrades.json"), r.upgrades); err != nil {
		t.Fatal(err)
	}
	if err := r.loadUpgrade(); err != nil || r.version != "1.3.8" {
		t.Fatal("launch overwrote later upgrade", err)
	}
	r.launch.AppliedRequest = "launch-second"
	r.launch.RuntimeVersion = "1.4"
	if err := r.loadUpgrade(); err != nil || r.version != "1.4" {
		t.Fatal(r.version, err)
	}
}

func TestLaunchRejectsInvalidSettings(t *testing.T) {
	current := launchConfig{ConfigVersion: 5}
	valid := map[string]any{"runtimeVersion": "1.4", "maxConcurrentDispatches": 2, "memoryMode": "platform"}
	for _, field := range []string{"capacity", "mode", "workspace", "version", "target"} {
		opts := make(map[string]any, len(valid))
		for k, v := range valid {
			opts[k] = v
		}
		q := aw.ExecutorRestart{RequestID: "next", LaunchConfigVersion: 6, TargetRuntimeVersion: "1.4"}
		switch field {
		case "capacity":
			opts["maxConcurrentDispatches"] = 51
		case "mode":
			opts["memoryMode"] = "unknown"
		case "workspace":
			opts["workspaceRoot"] = "./relative"
		case "version":
			q.LaunchConfigVersion = 4
		case "target":
			q.TargetRuntimeVersion = "other"
		}
		q.Launch, _ = json.Marshal(opts)
		if got, err := parseLaunch(&q, current); err == nil || got != current {
			t.Fatalf("%s accepted or changed current settings", field)
		}
	}
}
