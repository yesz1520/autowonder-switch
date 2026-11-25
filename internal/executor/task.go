package executor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yesz1520/autowonder-switch/internal/taskhooks"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

// A saved final reply allows redelivery without starting a second paid turn.
func cached(path string, out any) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, out)
}
func save(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writeFile(path, b)
}
func (r *runtime) task(ctx context.Context, s *aw.Session, q *aw.TaskDispatch) error {
	if q.DispatchID <= 0 {
		return errors.New("invalid dispatch ID")
	}
	dir := filepath.Join(r.dir, fmt.Sprintf("task-%d", q.DispatchID))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "result.json")
	var paused aw.TaskPaused
	if ok, e := cached(filepath.Join(dir, "paused.json"), &paused); e != nil {
		return e
	} else if ok {
		return r.send(&paused)
	}
	var prior aw.TaskResult
	if exists, err := r.loadFinal(path, &prior); err != nil {
		return err
	} else if exists {
		return r.send(&prior)
	}
	j, duplicate := r.reserve(r.tasks, q.DispatchID, 0, q.Attempt)
	ack := &aw.TaskAck{DispatchID: q.DispatchID}
	if duplicate {
		return s.Send(ctx, ack)
	}
	if j == nil {
		return s.Send(ctx, &aw.TaskBusy{DispatchID: q.DispatchID, Reason: "capacity exhausted"})
	}
	if err := s.Send(ctx, ack); err != nil {
		r.finish(r.tasks, q.DispatchID)
		return err
	}
	r.wg.Go(func() {
		defer r.finish(r.tasks, q.DispatchID)
		j.debug = newDispatchLog(q)
		result := aw.TaskResult{DispatchID: q.DispatchID, ResultProtocolVersion: 2, ResultID: aw.NewResultID()}
		marker := filepath.Join(dir, "started")
		prepared, ok, err := loadDelivery(dir)
		if err == nil && ok {
			j.packageHooks, err = taskhooks.Load(filepath.Join(dir, "package.zip"), dir, filepath.Join(dir, "workspace"), q.Attempt, r.codexBinary)
			if err == nil {
				err = r.deliver(j.ctx, q, dir, prepared, &result, j.packageHooks)
			}
		} else if err == nil {
			err = markStart(marker)
			if err == nil {
				err = r.execute(j, q, dir, &result)
			} else if errors.Is(err, errAlreadyStarted) {
				err = r.recoverTask(j, q, dir, &result)
			}
		}
		j.control.Lock()
		j.terminal = true
		paused := j.paused != nil
		j.control.Unlock()
		if j.ctx.Err() != nil && r.ctx != nil && r.ctx.Err() != nil {
			return
		}
		if paused {
			return
		}
		if err != nil {
			err = errors.Join(err, j.packageHooks.Run(j.ctx, "onFailure", "", nil))
		}
		err = errors.Join(err, j.packageHooks.Run(j.ctx, "cleanup", "", nil))
		if err == nil && q.WorkitemID > 0 {
			var proposal handoffProposal
			var exists bool
			exists, err = cached(filepath.Join(dir, "handoff-proposal.json"), &proposal)
			if err == nil && exists {
				if proposal.WorkitemID != q.WorkitemID {
					err = errors.New("handoff work item mismatch")
				} else {
					result.Handoff, err = json.Marshal(proposal.Target)
				}
			}
		}
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
		result.DebugLog = j.debug.upload(r.ctx, r.client, q, &result)
		r.finishDelivery(path, &result)
	})
	return nil
}
func (r *runtime) execute(j *job, q *aw.TaskDispatch, dir string, result *aw.TaskResult) error {
	file := filepath.Join(dir, "package.zip")
	if err := r.client.DownloadPackage(j.ctx, q, file); err != nil {
		return err
	}
	j.debug.event("package.verified", q.Size)
	workspace := filepath.Join(dir, "workspace")
	if err := r.ensureWorkspace(dir); err != nil {
		return err
	}
	var err error
	j.packageHooks, err = taskhooks.Load(file, dir, workspace, q.Attempt, r.codexBinary)
	if err != nil {
		return err
	}
	if err = j.packageHooks.Run(j.ctx, "beforeRepoPrepare", "", nil); err != nil {
		return err
	}
	repos, err := prepareRepos(j.ctx, file, dir, workspace, r.sources, q)
	if err != nil {
		return err
	}
	if err = j.packageHooks.Run(j.ctx, "afterRepoPrepare", "", nil); err != nil {
		return err
	}
	spec := harness.Request{MemoryMode: r.launch.MemoryMode, Platform: r.platform(q.DispatchMCPToken), Environment: q.EnvironmentVariables, Directory: workspace, Writable: len(repos) > 0}
	spec.Prompt, err = packageText(file, true, &spec)
	if err != nil {
		return err
	}
	if len(repos) > 0 {
		spec.Instructions = deliveryInstructions
		spec.OutputSchema = deliverySchema()
	}
	closeTools, err := r.openPackageTools(j.ctx, file, q.MCPSecrets, &spec)
	if err != nil {
		return err
	}
	defer closeTools()
	closeCapabilities, err := r.openPackageCapabilities(file, &spec)
	if err != nil {
		return err
	}
	defer closeCapabilities()
	if err = r.bindTaskMemory(q, &spec); err != nil {
		return err
	}
	if q.WorkitemID > 0 {
		spec.Platform = withHandoff(spec.Platform, dir, q.WorkitemID)
		spec.Platform = r.withAcceptance(spec.Platform, q.DispatchID)
		spec.Instructions += "\nIf the work item calls for a handoff, use executor.request_handoff to propose the next owner."
	}
	if q.ModelSelection != nil {
		spec.Model = q.ModelSelection.Model
	}
	if q.ResumeFromDispatchID != 0 {
		prior := filepath.Join(r.dir, fmt.Sprintf("task-%d", q.ResumeFromDispatchID), "session")
		b, err := os.ReadFile(prior)
		native := err == nil && q.ResumeProvider == r.provider() && q.ResumeSessionID != "" && string(b) == q.ResumeSessionID
		if native {
			spec.Session = string(b)
			spec.Fork = q.ResumeSessionBehavior == "FORK"
		} else if q.ResumeSessionBehavior == "FORK" {
			return errors.New("requested source harness session unavailable")
		}
		if len(repos) > 0 || !native || q.ResumeCheckpointURL != "" || q.ResumeCheckpointSHA256 != "" || q.ResumeCheckpointSeq > 0 || len(q.ResumeCheckpointCandidates) > 0 {
			priorText, err := r.restore(j.ctx, q, dir, repos)
			if err != nil {
				return err
			}
			spec.Prompt += "\n--- Previous dispatch ---\n" + priorText
		}

	}
	modelCtx, stop := context.WithCancel(j.ctx)
	defer stop()
	done := make(chan struct{})
	if err = j.packageHooks.Run(modelCtx, "beforeAgentStart", "", nil); err != nil {
		return err
	}
	commands, closeHooks, err := j.packageHooks.ToolCommands(modelCtx)
	if err != nil {
		return err
	}
	defer closeHooks()
	spec.ToolHookCommands = commands
	state := taskRecovery{PackageID: q.PackageID, Checksum: q.Checksum, ResultID: result.ResultID, Attempt: q.Attempt}
	for _, repo := range repos {
		state.Repositories = append(state.Repositories, preparedRepo{Name: repo.name, Base: repo.base, Head: repo.head, Remote: repo.remote, Branch: repo.branch, Verified: repo.verified, Spec: repo.spec})
	}
	if err = save(filepath.Join(dir, "task-recovery.json"), state); err != nil {
		return err
	}
	j.debug.event("model.started", int64(len(repos)))
	r.report(&aw.TaskProgress{DispatchID: q.DispatchID, EventType: "llm.started", Message: r.provider() + " started"})
	reply, err := r.runWorkflow(modelCtx, q, file, spec, harness.Hooks{TurnStarted: recordNativeTurn(filepath.Join(dir, "native-turn.json")), Session: func(native string) error {
		if err := writeFile(filepath.Join(dir, "session"), []byte(native)); err != nil {
			return err
		}
		j.control.Lock()
		j.native = native
		j.portable = true
		j.repos = repos
		j.stopModel = stop
		j.modelDone = done
		j.control.Unlock()
		return nil
	}, Ready: func(f func(context.Context, string) error) { r.mu.Lock(); j.steer = f; r.mu.Unlock() }, Message: func(text string) {
		j.control.Lock()
		j.latest = text
		j.debug.event("model.message.bytes", int64(len(text)))
		j.control.Unlock()
		r.report(&aw.TaskProgress{DispatchID: q.DispatchID, EventType: "llm.message", Message: text})
	}}, repos, j.packageHooks)
	err = errors.Join(err, j.packageHooks.Run(modelCtx, "afterAgentExit", "", nil))
	close(done)
	j.control.Lock()
	j.terminal = true
	wait := j.pauseFinished
	j.control.Unlock()
	if wait != nil {
		<-wait
		return context.Canceled
	}
	if err != nil {
		return err
	}
	if err = j.ctx.Err(); err != nil {
		return err
	}
	if err = save(filepath.Join(dir, "workflow-reply.json"), reply); err != nil {
		return err
	}
	return r.finishTask(j, q, dir, reply, repos, result)
}
func (r *runtime) finishTask(j *job, q *aw.TaskDispatch, dir string, reply harness.Reply, repos []checkout, result *aw.TaskResult) error {
	if err := writeFile(filepath.Join(dir, "session"), []byte(reply.Session)); err != nil {
		return err
	}
	r.mu.Lock()
	if reply.Model != "" {
		r.model = reply.Model
	}
	r.mu.Unlock()
	if len(repos) > 0 && j.packageHooks.Has("beforeCommit") {
		if err := markStart(filepath.Join(dir, "commit-hooks-started")); err != nil {
			return fmt.Errorf("commit hook delivery state: %w", err)
		}
	}
	prepared, err := prepareDelivery(j.ctx, dir, result.ResultID, reply.Session, reply.Text, repos, j.packageHooks)
	if err != nil {
		return err
	}
	prepared.PackageID = q.PackageID
	prepared.Checksum = q.Checksum
	snapshot, err := repoCheckpoint(j.ctx, r.provider(), q.DispatchID, q.Attempt, result.ResultID, reply.Session, prepared.Summary, repos)
	if err != nil {
		return err
	}
	if err = writeFile(filepath.Join(dir, "recovery.tar.gz"), snapshot); err != nil {
		return err
	}
	if err = save(filepath.Join(dir, "delivery.json"), prepared); err != nil {
		return err
	}
	return r.deliver(j.ctx, q, dir, prepared, result, j.packageHooks)
}
func (r *runtime) deliver(ctx context.Context, q *aw.TaskDispatch, dir string, d preparedDelivery, result *aw.TaskResult, h *taskhooks.Set) error {
	if d.PackageID != q.PackageID || d.Checksum != q.Checksum {
		return errors.New("dispatch package differs from saved delivery")
	}
	result.ResultID = d.ResultID
	summary := d.Summary
	var revisions []revision
	files := []aw.UploadFile{{Name: "result.md", Reader: strings.NewReader(summary)}}
	var pushed []string
	for i, p := range d.Repos {
		c := p.checkout(dir)
		if p.Intent.Mode == "push" {
			if !permitted(c.spec.AllowPush, q.AllowPush) || !permitted(c.spec.AllowCommit, q.AllowCommit) || !permitted(c.spec.AllowNetwork, q.AllowNetwork) {
				return errors.New("dispatch no longer permits publishing")
			}
			if err := c.publishWithHooks(ctx, h); err != nil {
				return fmt.Errorf("%s: %w (verified earlier in this delivery: %v)", c.name, err, pushed)
			}
			pushed = append(pushed, c.name+"@"+c.head)
			result.ResultSummary = "Verified remote commits: " + strings.Join(pushed, ", ")
			d.Repos[i].Verified = c.verified
			if err := save(filepath.Join(dir, "delivery.json"), d); err != nil {
				return err
			}
		}
		revisions = append(revisions, c.revision())
		patch, err := os.ReadFile(filepath.Join(dir, "change-"+p.Name+".patch"))
		if err != nil {
			return err
		}
		if len(patch) > 0 {
			files = append(files, aw.UploadFile{Name: p.Name + ".patch", Reader: bytes.NewReader(patch)})
		}
		if p.Intent.Mode != "patch" {
			summary += fmt.Sprintf("\n%s: %s %s %s", p.Name, p.Intent.Mode, c.branch, c.head)
		}
	}
	var metadata []json.RawMessage
	if len(revisions) > 0 {
		b, err := json.Marshal(map[string]any{"repositories": revisions})
		if err != nil {
			return err
		}
		files = append(files, aw.UploadFile{Name: "runtime-source-revision.json", Reader: bytes.NewReader(b)})
		metadata = make([]json.RawMessage, len(files))
		for i, f := range files {
			metadata[i], _ = json.Marshal(map[string]string{"path": "deliverables/" + f.Name})
		}
		files[0].Reader = strings.NewReader(summary)
	}
	if _, err := r.client.UploadArtifacts(ctx, q.DispatchID, files, result.ResultID, metadata); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, "recovery.tar.gz"))
	if err != nil {
		return err
	}
	ck, err := r.client.UploadCheckpoint(ctx, q.DispatchID, aw.CheckpointUpload{Seq: 1, Purpose: "COMPLETION", ResultID: result.ResultID, Provider: r.provider(), ProviderSessionID: d.Session, File: aw.UploadFile{Name: "checkpoint.tar.gz", Reader: bytes.NewReader(data)}})
	if err != nil {
		return err
	}
	if ck.SHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(data)) || ck.SizeBytes != int64(len(data)) {
		return errors.New("checkpoint bytes differ from receipt")
	}
	result.CheckpointSeq = ck.CheckpointSeq
	result.CheckpointSHA256 = ck.SHA256
	result.ResultSummary = summary
	return nil
}

// An interrupted process cannot know whether a turn already performed a tool
// call. Let the platform decide to retry rather than running it twice.
var errAlreadyStarted = errors.New("executor interrupted before saving this result")

func markStart(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return errAlreadyStarted
	}
	if err != nil {
		return err
	}
	return f.Close()
}
