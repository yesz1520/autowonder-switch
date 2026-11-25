package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

// The launch settings and request ID must become durable in the same write.
type restartState struct {
	History []string
	Launch  launchConfig
}

type launchConfig struct {
	RuntimeVersion string `json:"runtimeVersion"`
	WorkspaceRoot  string `json:"workspaceRoot"`
	DaemonName     string `json:"daemonName"`
	Debug          bool   `json:"debug"`
	MemoryMode     string `json:"memoryMode"`
	Capacity       int    `json:"maxConcurrentDispatches"`
	ConfigVersion  int64  `json:"configVersion"`
	AppliedRequest string `json:"appliedRequest"`
}

func (r *runtime) loadRestart() error {
	var state restartState
	ok, err := cached(filepath.Join(r.dir, "restart-state.json"), &state)
	if err != nil {
		return err
	}
	if !ok {
		if _, err = cached(filepath.Join(r.dir, "restarts.json"), &state.History); err != nil {
			return err
		}
	}
	r.restartHistory = state.History
	r.launch = state.Launch
	if n := len(state.History); n > 0 {
		r.restartID = state.History[n-1]
	}
	return nil
}

func parseLaunch(q *aw.ExecutorRestart, current launchConfig) (launchConfig, error) {
	if len(q.Launch) == 0 || string(q.Launch) == "null" {
		if q.TargetRuntimeVersion != "" || q.LaunchConfigVersion != 0 {
			return current, errors.New("missing launch configuration")
		}
		return current, nil
	}
	var next launchConfig
	if err := json.Unmarshal(q.Launch, &next); err != nil {
		return current, err
	}
	if q.LaunchConfigVersion <= 0 || q.LaunchConfigVersion < current.ConfigVersion || strings.TrimSpace(next.RuntimeVersion) == "" || next.RuntimeVersion != q.TargetRuntimeVersion {
		return current, errors.New("inconsistent launch configuration version")
	}
	if next.Capacity < 1 || next.Capacity > 50 {
		return current, errors.New("invalid launch capacity")
	}
	switch next.MemoryMode {
	case "none", "provider-local", "platform":
	default:
		return current, errors.New("invalid memory mode")
	}
	if strings.HasPrefix(next.WorkspaceRoot, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return current, err
		}
		next.WorkspaceRoot = filepath.Join(home, next.WorkspaceRoot[2:])
	}
	if next.WorkspaceRoot != "" && !filepath.IsAbs(next.WorkspaceRoot) {
		return current, errors.New("workspaceRoot must be absolute")
	}
	if next.WorkspaceRoot != "" {
		if err := os.MkdirAll(next.WorkspaceRoot, 0700); err != nil {
			return current, err
		}
		probe, err := os.MkdirTemp(next.WorkspaceRoot, ".autowonder-probe-")
		if err != nil {
			return current, err
		}
		if err = os.Remove(probe); err != nil {
			return current, err
		}
	}
	next.ConfigVersion = q.LaunchConfigVersion
	next.AppliedRequest = q.RequestID
	return next, nil
}

func (r *runtime) Handle(ctx context.Context, s *aw.Session, msg aw.Incoming) error {
	if r.launch.Debug {
		log.Printf("executor %d (%q) received %s", r.executorID, r.launch.DaemonName, msg.Kind())
	}
	return r.Endpoint.Handle(ctx, s, msg)
}

func (r *runtime) ensureWorkspace(dir string) error {
	link := filepath.Join(dir, "workspace")
	if r.launch.WorkspaceRoot == "" {
		return os.MkdirAll(link, 0700)
	}
	if info, err := os.Stat(link); err == nil {
		if !info.IsDir() {
			return errors.New("workspace path is not a directory")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	record := filepath.Join(dir, "workspace-location.json")
	var target string
	ok, err := cached(record, &target)
	if err != nil {
		return err
	}
	if !ok {
		target, err = os.MkdirTemp(r.launch.WorkspaceRoot, fmt.Sprintf("autowonder-%d-%s-", r.executorID, filepath.Base(dir)))
		if err != nil {
			return err
		}
		if err = save(record, target); err != nil {
			os.Remove(target)
			return err
		}
	}
	return os.Symlink(target, link)
}
