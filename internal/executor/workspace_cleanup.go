package executor

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func (r *runtime) queueCleanup(ctx context.Context, _ *aw.Session, q *aw.WorkitemWorkspaceCleanup) error {
	select {
	case r.cleanups <- q:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *runtime) runCleanups() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case q := <-r.cleanups:
			if err := r.cleanupWorkspaces(r.ctx, q); err != nil && r.ctx.Err() == nil {
				log.Printf("workspace cleanup: %v", err)
			}
		}
	}
}

func (r *runtime) cleanupWorkspaces(ctx context.Context, request *aw.WorkitemWorkspaceCleanup) error {
	if request.ExecutorID != r.executorID || request.TenantID <= 0 || request.WorkitemID <= 0 || request.PublishedAt <= 0 || request.WorkitemVersion < 0 {
		return nil
	}
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		suffix, ok := strings.CutPrefix(entry.Name(), "task-")
		if !ok || !entry.IsDir() {
			continue
		}
		id, err := strconv.ParseInt(suffix, 10, 64)
		if err != nil || id <= 0 || strconv.FormatInt(id, 10) != suffix {
			continue
		}
		r.mu.Lock()
		active := r.tasks[id] != nil
		r.mu.Unlock()
		if active {
			continue
		}
		if err := cleanDispatchWorkspace(filepath.Join(r.dir, entry.Name()), id, request); err != nil {
			log.Printf("workspace cleanup dispatch=%d: %v", id, err)
		}
	}
	return nil
}

func cleanDispatchWorkspace(dir string, id int64, q *aw.WorkitemWorkspaceCleanup) error {
	// A durable final result makes all future deliveries of this dispatch read-only.
	var result aw.TaskResult
	terminal, err := cached(filepath.Join(dir, "result.json"), &result)
	if err != nil || !terminal || result.DispatchID != id {
		return err
	}
	z, err := zip.OpenReader(filepath.Join(dir, "package.zip"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer z.Close()
	f, err := z.Open("manifest.json")
	if err != nil {
		return err
	}
	defer f.Close()
	var m struct{ DispatchID, ExecutorID, WorkitemID, TenantID, CreatedAt string }
	if err = json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	if m.DispatchID != strconv.FormatInt(id, 10) || m.ExecutorID != strconv.FormatInt(q.ExecutorID, 10) || m.WorkitemID != strconv.FormatInt(q.WorkitemID, 10) || m.TenantID != strconv.FormatInt(q.TenantID, 10) {
		return nil
	}
	// Both timestamps come from the platform. A reopened work item may already
	// have newer dispatches by the time an older cleanup frame reaches us.
	created, err := time.Parse(time.RFC3339Nano, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("package creation time: %w", err)
	}
	if created.UnixMilli() > q.PublishedAt {
		return nil
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == "workspace" && e.Type()&os.ModeSymlink != 0 {
			var recorded string
			ok, err := cached(filepath.Join(dir, "workspace-location.json"), &recorded)
			if err != nil {
				return err
			}
			actual, err := root.Readlink("workspace")
			if err != nil {
				return err
			}
			prefix := fmt.Sprintf("autowonder-%d-task-%d-", q.ExecutorID, id)
			if ok && actual == recorded && filepath.IsAbs(actual) && strings.HasPrefix(filepath.Base(actual), prefix) {
				if err = os.RemoveAll(actual); err != nil {
					return err
				}
			}
		}
		if e.Name() == "workspace" || strings.HasPrefix(e.Name(), "git-") {
			if err := root.RemoveAll(e.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}
