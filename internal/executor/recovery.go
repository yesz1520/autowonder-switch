package executor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
)

func restoreAnswer(path string, dispatch int64) (string, error) {
	_, entries, err := readCheckpoint(path, dispatch)
	return string(entries["result.md"]), err
}
func (r *runtime) restore(ctx context.Context, q *aw.TaskDispatch, dir string, repos []checkout) (string, error) {
	local := filepath.Join(r.dir, fmt.Sprintf("task-%d", q.ResumeFromDispatchID), "recovery.tar.gz")
	candidates := slices.Clone(q.ResumeCheckpointCandidates)
	if q.ResumeCheckpointURL != "" || q.ResumeCheckpointSHA256 != "" || q.ResumeCheckpointSeq > 0 {
		single := aw.CheckpointCandidate{DownloadURL: q.ResumeCheckpointURL, SHA256: q.ResumeCheckpointSHA256, CheckpointSeq: q.ResumeCheckpointSeq}
		if !slices.Contains(candidates, single) {
			candidates = append(candidates, single)
		}
	}
	if len(candidates) == 0 {
		return restoreRepos(ctx, local, q.ResumeFromDispatchID, repos)
	}
	path := filepath.Join(dir, "restore.tar.gz")
	err := errors.New("no portable checkpoint supplied")
	for _, v := range candidates {
		err = r.client.DownloadVerified(ctx, v.DownloadURL, v.SHA256, path)
		if err != nil && v.CheckpointSeq > 0 {
			if link, e := r.client.RefreshCheckpoint(ctx, q.DispatchID, v.CheckpointSeq); e == nil && link.SHA256 == v.SHA256 {
				err = r.client.DownloadVerified(ctx, link.DownloadURL, v.SHA256, path)
			}
		}
		if err == nil {
			var text string
			text, err = restoreRepos(ctx, path, q.ResumeFromDispatchID, repos)
			if err == nil {
				return text, nil
			}
		}
	}
	// A local copy may outlive its signed URL. Only use a snapshot selected by the platform.
	if f, openErr := os.Open(local); openErr == nil {
		defer f.Close()
		h := sha256.New()
		n, hashErr := io.Copy(h, io.LimitReader(f, archiveLimit+1))
		if hashErr == nil && n <= archiveLimit {
			digest := fmt.Sprintf("%x", h.Sum(nil))
			for _, v := range candidates {
				if strings.EqualFold(strings.TrimPrefix(v.SHA256, "sha256:"), digest) {
					return restoreRepos(ctx, local, q.ResumeFromDispatchID, repos)
				}
			}
		}
	}
	return "", err
}
func (r *runtime) pause(ctx context.Context, s *aw.Session, q *aw.TaskPause) error {
	r.mu.Lock()
	j := r.tasks[q.DispatchID]
	r.mu.Unlock()
	if j == nil {
		if q.Cancel == nil || !*q.Cancel {
			var paused aw.TaskPaused
			ok, err := cached(filepath.Join(r.dir, fmt.Sprintf("task-%d", q.DispatchID), "paused.json"), &paused)
			if err != nil {
				return err
			}
			if ok {
				return s.Send(ctx, &paused)
			}
		}
		return s.Send(ctx, &aw.TaskPauseFailed{DispatchID: q.DispatchID, Error: "dispatch is not running"})
	}
	if q.Cancel != nil && *q.Cancel {
		j.cancel()
		return nil
	}
	r.wg.Go(func() {
		j.control.Lock()
		defer j.control.Unlock()
		if j.pauseFinished != nil && j.paused == nil {
			return
		}
		if j.paused != nil {
			r.report(j.paused)
			return
		}
		if j.terminal || j.native == "" || !j.portable {
			r.report(&aw.TaskPauseFailed{DispatchID: q.DispatchID, Error: "no active portable Codex checkpoint"})
			return
		}
		if j.stopModel != nil {
			finished := make(chan struct{})
			j.pauseFinished = finished
			defer close(finished)
			j.stopModel()
			done := j.modelDone
			j.control.Unlock()
			<-done
			j.control.Lock()
		}
		latest := j.latest
		if latest == "" {
			latest = "The preceding turn was paused before its first message. Continue from the work item."
		}
		data, err := repoCheckpoint(j.ctx, q.DispatchID, j.attempt, "", j.native, latest, j.repos)
		if err == nil {
			err = writeFile(filepath.Join(r.dir, fmt.Sprintf("task-%d", q.DispatchID), "recovery.tar.gz"), data)
		}
		var ck aw.CheckpointReceipt
		if err == nil {
			upload, stop := context.WithTimeout(j.ctx, 30*time.Second)
			defer stop()
			for attempt := 0; attempt < 3; attempt++ {
				ck, err = r.client.UploadCheckpoint(upload, q.DispatchID, aw.CheckpointUpload{Seq: 1, Purpose: "RECOVERY", Provider: "codex", ProviderSessionID: j.native, File: aw.UploadFile{Name: "checkpoint.tar.gz", Reader: bytes.NewReader(data)}})
				var status *aw.HTTPError
				if err == nil || upload.Err() != nil || !errors.As(err, &status) || (status.StatusCode < 500 && status.StatusCode != 429 && status.StatusCode != 408) {
					break
				}
				if attempt < 2 {
					select {
					case <-upload.Done():
					case <-time.After(time.Duration(attempt+1) * 250 * time.Millisecond):
					}
				}
			}
		}
		if err == nil && (ck.SHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(data)) || ck.SizeBytes != int64(len(data))) {
			err = errors.New("invalid recovery checkpoint receipt")
		}
		paused := &aw.TaskPaused{DispatchID: q.DispatchID, CheckpointSeq: ck.CheckpointSeq, CheckpointSHA256: ck.SHA256}
		if err == nil {
			err = save(filepath.Join(r.dir, fmt.Sprintf("task-%d", q.DispatchID), "paused.json"), paused)
		}
		if err != nil {
			r.report(&aw.TaskPauseFailed{DispatchID: q.DispatchID, Error: err.Error()})
			return
		}
		j.paused = paused
		j.cancel()
		r.report(paused)
	})
	return nil
}

// Keep paused.json as the idempotent reply to a repeated dispatch, and track the
// pending pause separately until the server confirms the checkpoint identity.
func (r *runtime) confirmPause(_ context.Context, _ *aw.Session, a *aw.TaskResultAck) error {
	if !a.Accepted || a.PauseCheckpointConfirmed == nil || !*a.PauseCheckpointConfirmed {
		return nil
	}
	dir := filepath.Join(r.dir, fmt.Sprintf("task-%d", a.DispatchID))
	p := filepath.Join(dir, "paused.json")
	var m aw.TaskPaused
	ok, err := cached(p, &m)
	if err != nil || !ok {
		return err
	}
	if a.CheckpointSHA256 != m.CheckpointSHA256 || a.CheckpointSeq != m.CheckpointSeq {
		return nil
	}
	return writeFile(filepath.Join(dir, "pause-confirmed"), nil)
}
func (r *runtime) replayPauses(ctx context.Context, s *aw.Session) error {
	paths, err := filepath.Glob(filepath.Join(r.dir, "task-*", "paused.json"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		if _, err = os.Stat(filepath.Join(filepath.Dir(p), "pause-confirmed")); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		var m aw.TaskPaused
		if _, err = cached(p, &m); err != nil {
			return err
		}
		if err = s.Send(ctx, &m); err != nil {
			return err
		}
	}
	return nil
}
