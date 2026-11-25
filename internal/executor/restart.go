package executor

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

var ErrRestartRequested = errors.New("executor restart requested")

func (r *runtime) restart(ctx context.Context, s *aw.Session, q *aw.ExecutorRestart) error {
	result := func(status, msg string) error {
		return s.Send(ctx, &aw.ExecutorRestartResult{RequestID: q.RequestID, Status: status, Message: msg})
	}
	if !r.files.RestartEnabled || q.RequestID == "" || (q.ExecutorID != 0 && q.ExecutorID != r.executorID) {
		return result("FAILED", "restart unavailable or invalid identity")
	}
	if q.Update != nil && *q.Update {
		return result("FAILED", "binary updates are unsupported")
	}
	if q.RestartMode != "" && q.RestartMode != "WHEN_IDLE" && q.RestartMode != "FORCE" {
		return result("FAILED", "unknown restart mode")
	}
	if q.PreviousStartedAt > time.Now().UnixMilli() {
		return result("FAILED", "previous start time is in the future")
	}
	r.mu.Lock()
	if r.draining == q.RequestID {
		r.mu.Unlock()
		return nil
	}
	if r.draining != "" {
		r.mu.Unlock()
		return result("FAILED", "another restart is pending")
	}
	if slices.Contains(r.restartHistory, q.RequestID) {
		r.mu.Unlock()
		return nil
	}
	launch, err := parseLaunch(q, r.launch)
	if err != nil {
		r.mu.Unlock()
		return result("FAILED", err.Error())
	}
	next := append(slices.Clone(r.restartHistory), q.RequestID)
	if err := save(filepath.Join(r.dir, "restart-state.json"), restartState{History: next, Launch: launch}); err != nil {
		r.mu.Unlock()
		return result("FAILED", err.Error())
	}
	r.restartHistory = next
	r.draining = q.RequestID
	var idle chan struct{}
	if q.RestartMode == "WHEN_IDLE" && len(r.tasks)+len(r.turns) > 0 {
		idle = make(chan struct{})
		r.idle = idle
	}
	r.mu.Unlock()
	r.wg.Go(func() {
		msg := func(status string) { r.report(&aw.ExecutorRestartResult{RequestID: q.RequestID, Status: status}) }
		if idle != nil {
			tick := time.NewTicker(time.Minute)
			defer tick.Stop()
			msg("WAITING_IDLE")
		wait:
			for {
				select {
				case <-r.ctx.Done():
					return
				case <-idle:
					break wait
				case <-tick.C:
					msg("WAITING_IDLE")
				}
			}
		}
		if delay := time.Until(time.UnixMilli(q.PreviousStartedAt + 1)); delay > 0 {
			select {
			case <-r.ctx.Done():
				return
			case <-time.After(delay):
			}
		}
		msg("RESTARTING")
		r.shutdown(ErrRestartRequested)
	})
	return nil
}
