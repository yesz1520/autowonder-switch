package executor

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func TestWorkspaceCleanupIsolation(t *testing.T) {
	dir := t.TempDir()
	r := runtime{dir: dir, executorID: 7, tasks: map[int64]*job{5: {}}}
	r.launch.WorkspaceRoot = t.TempDir()
	published := time.Now().UTC()
	for id := int64(1); id <= 7; id++ {
		base := filepath.Join(dir, fmt.Sprintf("task-%d", id))
		if err := r.ensureWorkspace(base); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"git-project"} {
			if err := os.MkdirAll(filepath.Join(base, name), 0700); err != nil {
				t.Fatal(err)
			}
		}
		if id != 6 {
			if err := save(filepath.Join(base, "result.json"), aw.TaskResult{DispatchID: id, Success: true}); err != nil {
				t.Fatal(err)
			}
		}
		f, err := os.Create(filepath.Join(base, "package.zip"))
		if err != nil {
			t.Fatal(err)
		}
		z := zip.NewWriter(f)
		w, err := z.Create("manifest.json")
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{"tenantId": "3", "executorId": "7", "workitemId": "91", "dispatchId": fmt.Sprint(id), "createdAt": published.Add(-time.Hour).Format(time.RFC3339Nano)}
		switch id {
		case 2:
			m["tenantId"] = "4"
		case 3:
			m["workitemId"] = "92"
		case 4:
			m["createdAt"] = published.Add(time.Second).Format(time.RFC3339Nano)
		case 7:
			m["executorId"] = "8"
		}
		if err = json.NewEncoder(w).Encode(m); err != nil {
			t.Fatal(err)
		}
		if err = z.Close(); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	q := aw.WorkitemWorkspaceCleanup{ExecutorID: 7, TenantID: 3, WorkitemID: 91, WorkitemVersion: 4, PublishedAt: published.UnixMilli()}
	for range 2 {
		if err := r.cleanupWorkspaces(t.Context(), &q); err != nil {
			t.Fatal(err)
		}
	}
	targets, err := os.ReadDir(r.launch.WorkspaceRoot)
	if err != nil || len(targets) != 6 {
		t.Fatalf("external targets %v %v", targets, err)
	}
	for id := int64(1); id <= 7; id++ {
		for _, name := range []string{"workspace", "git-project", "package.zip", "result.json"} {
			_, err := os.Stat(filepath.Join(dir, fmt.Sprintf("task-%d", id), name))
			removed := (id == 1 && (name == "workspace" || name == "git-project")) || (id == 6 && name == "result.json")
			if (removed && !os.IsNotExist(err)) || (!removed && err != nil) {
				t.Errorf("dispatch %d %s: %v", id, name, err)
			}
		}
	}
}
