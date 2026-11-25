package executor

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

// The version and the unsent result change together. A crash between recording
// the version and queuing its receipt must not lose the result.
type upgradeState struct {
	VersionRevision string
	LaunchRequest   string
	Version         string
	Pending         string
	Results         map[string]aw.ExecutorUpgradeResult
}

// RecordedVersion returns the last platform version, unless a newer local edit is waiting to apply.
func RecordedVersion(dir, base string, id int64, revision string) string {
	var record struct{ Version, VersionRevision string }
	ok, err := cached(filepath.Join(identityDir(dir, base, id), "upgrades.json"), &record)
	if !ok || err != nil || record.VersionRevision != revision {
		return ""
	}
	return record.Version
}

func (r *runtime) loadUpgrade() error {
	nativeVersion := r.version
	ok, err := cached(filepath.Join(r.dir, "upgrades.json"), &r.upgrades)
	if err != nil {
		return err
	}
	if ok {
		if r.upgrades.Results == nil {
			return errors.New("invalid upgrade history")
		}
		if r.upgrades.Version != "" {
			r.version = r.upgrades.Version
		}
	} else {
		r.upgrades.Results = make(map[string]aw.ExecutorUpgradeResult)
	}
	if r.launch.AppliedRequest != "" && r.upgrades.LaunchRequest != r.launch.AppliedRequest {
		r.upgrades.LaunchRequest = r.launch.AppliedRequest
		r.upgrades.Version = r.launch.RuntimeVersion
		if err = save(filepath.Join(r.dir, "upgrades.json"), r.upgrades); err != nil {
			return err
		}
		r.version = r.upgrades.Version
	}
	if r.files.VersionRevision != "" && r.upgrades.VersionRevision != r.files.VersionRevision {
		r.upgrades.VersionRevision = r.files.VersionRevision
		r.upgrades.Version = r.files.ReportedVersion
		r.version = nativeVersion
		if r.upgrades.Version != "" {
			r.version = r.upgrades.Version
		}
		return save(filepath.Join(r.dir, "upgrades.json"), r.upgrades)
	}
	return nil
}

func (r *runtime) flushUpgrade(ctx context.Context, s *aw.Session) error {
	id := r.upgrades.Pending
	if id == "" {
		return nil
	}
	result, ok := r.upgrades.Results[id]
	if !ok {
		return errors.New("pending upgrade result is missing")
	}
	if err := r.Endpoint.Send(ctx, s, &result); err != nil {
		return err
	}
	next := r.upgrades
	next.Pending = ""
	if err := save(filepath.Join(r.dir, "upgrades.json"), next); err != nil {
		return err
	}
	r.upgrades = next
	return nil
}

func (r *runtime) upgrade(ctx context.Context, s *aw.Session, q *aw.ExecutorUpgrade) error {
	if q.RequestID == "" || (q.ExecutorID != 0 && q.ExecutorID != r.executorID) {
		return s.Send(ctx, &aw.ExecutorUpgradeResult{RequestID: q.RequestID, Phase: "failed", Error: "invalid upgrade request identity"})
	}
	if err := r.flushUpgrade(ctx, s); err != nil {
		return err
	}
	if previous, ok := r.upgrades.Results[q.RequestID]; ok {
		return r.Endpoint.Send(ctx, s, &previous)
	}
	result := aw.ExecutorUpgradeResult{RequestID: q.RequestID, Phase: "success", CurrentVersion: q.TargetVersion}
	if q.ResultReceiptAck != nil && *q.ResultReceiptAck {
		result.ReceiptID = aw.NewResultID()
	}
	next := r.upgrades
	if strings.TrimSpace(q.TargetVersion) == "" {
		terminal := true
		result.Phase, result.Error, result.Terminal, result.CurrentVersion = "failed", "missing targetVersion", &terminal, ""
	} else {
		next.Version = q.TargetVersion
	}
	next.Pending = q.RequestID
	next.Results[q.RequestID] = result
	if err := save(filepath.Join(r.dir, "upgrades.json"), next); err != nil {
		delete(next.Results, q.RequestID)
		return err
	}
	r.upgrades = next
	if result.Phase == "success" {
		r.mu.Lock()
		r.version = next.Version
		r.mu.Unlock()
	}
	return r.flushUpgrade(ctx, s)
}
