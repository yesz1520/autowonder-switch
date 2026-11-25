package executor

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
	"github.com/yesz1520/autowonder-switch/internal/taskhooks"
)

type taskRecovery struct {
	PackageID, Checksum, ResultID string
	Attempt                       int64
	Repositories                  []preparedRepo
}

func (r *runtime) recoverTask(j *job, q *aw.TaskDispatch, dir string, result *aw.TaskResult) error {
	var state taskRecovery
	ok, err := cached(filepath.Join(dir, "task-recovery.json"), &state)
	if err != nil {
		return err
	}
	if !ok || state.PackageID != q.PackageID || state.Checksum != q.Checksum || state.Attempt != q.Attempt || state.ResultID == "" {
		return errAlreadyStarted
	}
	result.ResultID = state.ResultID
	work := filepath.Join(dir, "workspace")
	pkg := filepath.Join(dir, "package.zip")
	j.packageHooks, err = taskhooks.Load(pkg, dir, work, q.Attempt, r.codexBinary)
	if err != nil {
		return err
	}
	var reply harness.Reply
	ok, err = cached(filepath.Join(dir, "workflow-reply.json"), &reply)
	if err != nil {
		return err
	}
	if !ok {
		// A completed native turn may still be waiting for workflow approval or an exit hook.
		steps, err := gatedSteps(pkg, j.packageHooks.Has("beforeStep") || j.packageHooks.Has("afterStep"))
		if err != nil {
			return err
		}
		if len(steps) > 0 || j.packageHooks.Has("afterAgentExit") {
			return errors.New("interrupted before workflow completion was recorded")
		}
		reader, supported := r.backend.(harness.TurnReader)
		var native nativeTurn
		found, err := cached(filepath.Join(dir, "native-turn.json"), &native)
		if err != nil {
			return err
		}
		if !found || !supported {
			return errAlreadyStarted
		}
		var status string
		reply, status, err = reader.ReadTurn(j.ctx, work, native.Session, native.Turn)
		if err != nil {
			return err
		}
		if status != "completed" || strings.TrimSpace(reply.Text) == "" || reply.Session != native.Session {
			return fmt.Errorf("no completed Codex reply for previous task turn (status %s)", status)
		}
		if err = save(filepath.Join(dir, "workflow-reply.json"), reply); err != nil {
			return err
		}
	}
	var repos []checkout
	for _, repo := range state.Repositories {
		c := repo.checkout(dir)
		c.spec.AllowCommit = permitted(c.spec.AllowCommit, q.AllowCommit)
		c.spec.AllowPush = permitted(c.spec.AllowPush, q.AllowPush)
		c.spec.AllowNetwork = permitted(c.spec.AllowNetwork, q.AllowNetwork)
		repos = append(repos, c)
	}
	return r.finishTask(j, q, dir, reply, repos, result)
}
